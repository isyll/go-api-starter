package event

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"

	"github.com/isyll/go-grpc-starter/pkg/logger"
)

type Worker struct {
	server *asynq.Server
	mux    *asynq.ServeMux
	proc   *Processor
	logger *logger.Logger
}

// Queues are namespaced so each Asynq server only consumes its own work.
// Without the prefix, the email and notification servers would pick up event
// tasks they have no handler for and burn retries bouncing them around.
const (
	QueueHigh   = "events:high"
	QueueNormal = "events:normal"
	QueueLow    = "events:low"
)

func QueueNames() []string {
	return []string{QueueHigh, QueueNormal, QueueLow}
}

type WorkerConfig struct {
	Concurrency int
	Queues      map[string]int
}

func DefaultWorkerConfig() WorkerConfig {
	return WorkerConfig{
		Concurrency: 10,
		Queues: map[string]int{
			QueueHigh:   6,
			QueueNormal: 3,
			QueueLow:    1,
		},
	}
}

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

func (w *Worker) Start() error {
	w.mux.HandleFunc(TaskTypeDispatch, w.proc.ProcessTask)
	w.logger.Info("event-dispatcher worker starting")
	return w.server.Start(w.mux)
}

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
