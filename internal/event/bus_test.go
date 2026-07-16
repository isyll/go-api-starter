package event

import (
	"context"
	"errors"
	"testing"

	"github.com/hibiken/asynq"

	"github.com/isyll/go-grpc-starter/pkg/logger"
)

type testEvent struct {
	CommonFields
	Token string
}

func (*testEvent) EventType() string { return "test.unit" }

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
