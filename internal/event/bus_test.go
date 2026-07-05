package event

import (
	"context"
	"errors"
	"testing"

	"github.com/hibiken/asynq"

	"github.com/isyll/go-grpc-starter/pkg/logger"
)

// testEvent is registered on the bus via Subscribe[*testEvent].
type testEvent struct {
	CommonFields
	Token string
}

func (*testEvent) EventType() string { return "test.unit" }

// A sync handler registered with WithCritical() must surface its error.
func TestPublishCriticalSurfacesError(t *testing.T) {
	logx := logger.New("test")
	bus := New(nil, logx)

	wantErr := errors.New("audit log unavailable")
	Subscribe(
		bus,
		func(_ context.Context, _ *testEvent) error {
			return wantErr
		},
		WithCritical(),
	)

	gotErr := bus.Publish(
		context.Background(),
		&testEvent{Token: "x"},
	)
	if !errors.Is(gotErr, wantErr) {
		t.Fatalf(
			"Publish returned %v; want critical handler error %v",
			gotErr, wantErr,
		)
	}
}

// A non-critical sync handler failure must never reach the publisher.
func TestPublishNonCriticalSwallowsError(t *testing.T) {
	logx := logger.New("test")
	bus := New(nil, logx)

	Subscribe(
		bus,
		func(_ context.Context, _ *testEvent) error {
			return errors.New("redis pipeline failed")
		},
	)

	if err := bus.Publish(
		context.Background(),
		&testEvent{Token: "y"},
	); err != nil {
		t.Fatalf(
			"non-critical handler error leaked through Publish: %v",
			err,
		)
	}
}

// A panicking critical handler is recovered and surfaced as an error.
func TestPublishHandlerPanicIsRecovered(t *testing.T) {
	logx := logger.New("test")
	bus := New(nil, logx)

	Subscribe(
		bus,
		func(_ context.Context, _ *testEvent) error {
			panic("boom")
		},
		WithCritical(),
	)

	err := bus.Publish(
		context.Background(),
		&testEvent{Token: "z"},
	)
	if err == nil {
		t.Fatal(
			"panic in critical sync handler did not surface as error",
		)
	}
}

type stubDispatcher struct {
	err   error
	calls int
}

func (d *stubDispatcher) Enqueue(context.Context, Event, []asynq.Option) error {
	d.calls++
	return d.err
}

// A redelivered event whose task already exists must not count as failure.
func TestPublishDuplicateEnqueueIsSuccess(t *testing.T) {
	logx := logger.New("test")

	for _, dupErr := range []error{asynq.ErrDuplicateTask, asynq.ErrTaskIDConflict} {
		disp := &stubDispatcher{err: dupErr}
		bus := New(disp, logx)

		SubscribeAsync(
			bus,
			func(_ context.Context, _ *testEvent) error { return nil },
			WithTaskIDFn(func(Event) string { return "fixed-id" }),
		)

		if err := bus.Publish(context.Background(), &testEvent{Token: "x"}); err != nil {
			t.Fatalf("Publish returned %v for %v; want nil", err, dupErr)
		}
		if disp.calls != 1 {
			t.Fatalf("dispatcher called %d times; want 1", disp.calls)
		}
	}
}

// Real enqueue failures surface so the outbox marks the row for retry.
func TestPublishEnqueueFailureSurfaces(t *testing.T) {
	logx := logger.New("test")
	disp := &stubDispatcher{err: errors.New("redis down")}
	bus := New(disp, logx)

	SubscribeAsync(
		bus,
		func(_ context.Context, _ *testEvent) error { return nil },
		WithTaskIDFn(func(Event) string { return "fixed-id" }),
	)

	err := bus.Publish(context.Background(), &testEvent{Token: "x"})
	if !errors.Is(err, ErrEnqueueFailed) {
		t.Fatalf("Publish returned %v; want ErrEnqueueFailed", err)
	}
}
