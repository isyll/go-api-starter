package emails

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"

	"github.com/isyll/go-grpc-starter/pkg/config"
	"github.com/isyll/go-grpc-starter/pkg/logger"
)

type Worker struct {
	server    *asynq.Server
	mux       *asynq.ServeMux
	processor *Processor
	cfg       *config.EmailConfig
	logger    *logger.Logger
}

func NewWorker(
	redisAddr string,
	redisPassword string,
	processor *Processor,
	cfg *config.EmailConfig,
	logx *logger.Logger,
) *Worker {
	server := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: redisPassword,
		},
		asynq.Config{
			Concurrency:     cfg.Email.Worker.Concurrency,
			ShutdownTimeout: 30 * time.Second,
			Queues: map[string]int{
				string(PriorityHigh):   6,
				string(PriorityNormal): 3,
				string(PriorityLow):    1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(
				func(_ context.Context, task *asynq.Task, err error) {
					logx.Error("Email task failed",
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
	w.mux.HandleFunc(TaskSendEmail, w.processor.ProcessTask)
	w.mux.HandleFunc(TaskScheduledEmail, w.processor.ProcessTask)
	w.mux.HandleFunc(TaskBulkEmail, w.processor.ProcessTask)

	w.logger.Info("Starting email worker",
		"concurrency", w.cfg.Email.Worker.Concurrency,
	)

	return w.server.Start(w.mux)
}

func (w *Worker) Run() error {
	if err := w.Start(); err != nil {
		return fmt.Errorf("failed to start email worker: %w", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	w.logger.Info("Shutting down email worker...")
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
