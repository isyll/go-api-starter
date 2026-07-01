package events

import (
	"context"
	"errors"
	"testing"

	"github.com/isyll/go-api-starter/pkg/logger"
)

// testEvent is a sync-only event with no factory; we
// register the type on the bus by Subscribe[*testEvent].
type testEvent struct {
	CommonFields
	Token string
}

func (*testEvent) EventType() string { return "test.unit" }

// TestPublishCriticalSurfacesError verifies that a sync
// handler registered with WithCritical() returns its
// error through Publish — the only seam any future
// "must succeed or abort" handler depends on. Without
// this test the seam silently rots: today's handlers all
// log-and-swallow.
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

// TestPublishNonCriticalSwallowsError verifies the inverse
// invariant: a non-critical sync handler's failure must
// NEVER surface to the publisher (cache invalidation is
// the canonical example; Redis being unreachable cannot
// fail the originating request).
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

// TestPublishHandlerPanicIsRecovered verifies that a
// panicking sync handler does not propagate; the bus
// recovers, logs, and returns a non-nil error when the
// handler was registered WithCritical.
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
