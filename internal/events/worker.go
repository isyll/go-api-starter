package events

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"

	"github.com/isyll/go-api-starter/pkg/logger"
)

// Worker hosts the events:dispatch queue. Every payload
// is one envelope; routing happens inside the bus once
// decoded.
type Worker struct {
	server *asynq.Server
	mux    *asynq.ServeMux
	proc   *Processor
	logger *logger.Logger
}

// WorkerConfig captures the tunables for the events:dispatch
// Asynq server.
type WorkerConfig struct {
	// Concurrency is the maximum number of tasks processed in
	// parallel by this worker process.
	Concurrency int
	// Queues maps Asynq queue names to weights (priority).
	Queues map[string]int
	// RetryMax caps the per-task retry budget.
	RetryMax int
}

// DefaultWorkerConfig returns the recommended defaults: ten worker
// goroutines split across high/normal/low queues, with five retries.
func DefaultWorkerConfig() WorkerConfig {
	return WorkerConfig{
		Concurrency: 10,
		Queues: map[string]int{
			"high":   6,
			"normal": 3,
			"low":    1,
		},
		RetryMax: 5,
	}
}

// NewWorker builds the events:dispatch Asynq worker bound to the
// supplied Bus. The Bus passed in MUST be configured without an
// AsyncDispatcher so handlers cannot re-publish from the worker.
func NewWorker(
	redisAddr string,
	redisPassword string,
	bus *Bus,
	cfg WorkerConfig,
	logx *logger.Logger,
) *Worker {
	server := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: redisPassword,
		},
		asynq.Config{
			Concurrency:     cfg.Concurrency,
			Queues:          cfg.Queues,
			ShutdownTimeout: 30 * time.Second,
			ErrorHandler: asynq.ErrorHandlerFunc(
				func(_ context.Context, t *asynq.Task, err error) {
					logx.Error(
						"event task failed",
						"type", t.Type(),
						"error", err,
					)
				},
			),
			Logger: &asynqLogger{logger: logx},
		},
	)

	return &Worker{
		server: server,
		mux:    asynq.NewServeMux(),
		proc:   NewProcessor(bus, logx),
		logger: logx,
	}
}

// Start registers the Processor on the events:dispatch task type
// and begins consuming from Redis. Non-blocking; pair with Run for
// blocking lifecycle management.
func (w *Worker) Start() error {
	w.mux.HandleFunc(TaskTypeDispatch, w.proc.ProcessTask)
	w.logger.Info("event-dispatcher worker starting")
	return w.server.Start(w.mux)
}

// Run starts the worker and blocks until SIGINT/SIGTERM is received,
// then gracefully shuts down.
func (w *Worker) Run() error {
	if err := w.Start(); err != nil {
		return fmt.Errorf(
			"event-dispatcher worker start: %w", err,
		)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	w.logger.Info("event-dispatcher worker shutting down")
	w.server.Shutdown()
	return nil
}

// Shutdown stops the worker, draining in-flight tasks within the
// configured ShutdownTimeout.
func (w *Worker) Shutdown() {
	w.server.Shutdown()
}

type asynqLogger struct {
	logger *logger.Logger
}

func (l *asynqLogger) Debug(args ...any) {
	l.logger.Debug(fmt.Sprint(args...))
}

func (l *asynqLogger) Info(args ...any) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *asynqLogger) Warn(args ...any) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *asynqLogger) Error(args ...any) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *asynqLogger) Fatal(args ...any) {
	l.logger.Fatal(fmt.Sprint(args...))
}
