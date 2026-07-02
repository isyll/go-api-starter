package notifications

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"

	"github.com/isyll/go-api-starter/pkg/config"
	"github.com/isyll/go-api-starter/pkg/logger"
)

type Worker struct {
	server    *asynq.Server
	mux       *asynq.ServeMux
	processor *Processor
	cfg       *config.NotificationsConfig
	logger    *logger.Logger
}

func NewWorker(
	redisAddr string,
	redisPassword string,
	processor *Processor,
	cfg *config.NotificationsConfig,
	logx *logger.Logger,
) *Worker {
	server := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: redisPassword,
		},
		asynq.Config{
			Concurrency:     cfg.Worker.Concurrency,
			ShutdownTimeout: 30 * time.Second,
			Queues: map[string]int{
				string(PriorityHigh):   6,
				string(PriorityNormal): 3,
				string(PriorityLow):    1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(
				func(ctx context.Context, task *asynq.Task, err error) {
					logx.Error("Task failed",
						"task", task.Type(),
						"error", err,
					)
				},
			),
			Logger: &asynqLogger{logger: logx},
		},
	)

	mux := asynq.NewServeMux()

	return &Worker{
		server:    server,
		mux:       mux,
		processor: processor,
		cfg:       cfg,
		logger:    logx,
	}
}

func (w *Worker) Start() error {
	w.mux.HandleFunc(TaskSendNotification, w.processor.ProcessTask)
	w.mux.HandleFunc(TaskScheduleNotification, w.processor.ProcessTask)

	w.logger.Info("Starting notification worker",
		"concurrency", w.cfg.Worker.Concurrency,
	)

	return w.server.Start(w.mux)
}

func (w *Worker) Run() error {
	if err := w.Start(); err != nil {
		return fmt.Errorf("failed to start worker: %w", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	w.logger.Info("Shutting down notification worker...")
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
