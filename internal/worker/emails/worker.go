package emails

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

// Worker wraps an Asynq server with a mux pre-wired for the email
// queues (high, normal, low).
type Worker struct {
	server    *asynq.Server
	mux       *asynq.ServeMux
	processor *Processor
	cfg       *config.EmailConfig
	logger    *logger.Logger
}

// NewWorker constructs a Worker around the given Processor and starts
// the Asynq server listening on the email queues from config.
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
				func(ctx context.Context, task *asynq.Task, err error) {
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

// Start begins processing email tasks
func (w *Worker) Start() error {
	w.mux.HandleFunc(TaskSendEmail, w.processor.ProcessTask)
	w.mux.HandleFunc(TaskScheduledEmail, w.processor.ProcessTask)
	w.mux.HandleFunc(TaskBulkEmail, w.processor.ProcessTask)

	w.logger.Info("Starting email worker",
		"concurrency", w.cfg.Email.Worker.Concurrency,
	)

	return w.server.Start(w.mux)
}

// Run starts the worker and blocks until shutdown signal
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

// Shutdown gracefully stops the worker
func (w *Worker) Shutdown() {
	w.server.Shutdown()
}

// asynqLogger adapts our logger to asynq's logger interface
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
