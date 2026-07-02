// Command email_sender is the standalone Asynq worker binary
package main

import (
	"log"

	"github.com/isyll/go-api-starter/internal/worker/emails"
	"github.com/isyll/go-api-starter/pkg/config"
	appenv "github.com/isyll/go-api-starter/pkg/env"
	"github.com/isyll/go-api-starter/pkg/locale"
	"github.com/isyll/go-api-starter/pkg/logger"
)

func main() {
	env := appenv.InitApp()

	cfg, err := config.LoadAllConfigs()
	if err != nil {
		log.Fatal("Failed to load configs", "error", err)
	}

	logx := logger.New(env)
	logx.Info("Starting email worker", "env", env)

	redisHost, redisPassword := cfg.Redis.Credentials()

	localizer, err := locale.New(cfg.App)
	if err != nil {
		logx.Fatal("i18n initialization failed: %w", err)
	}

	processor := emails.NewProcessor(cfg.Email, logx, localizer)

	worker := emails.NewWorker(
		redisHost,
		redisPassword,
		processor,
		cfg.Email,
		logx,
	)

	logx.Info("Email worker starting",
		"redis", redisHost,
		"concurrency", cfg.Email.Email.Worker.Concurrency,
		"provider", cfg.Email.Email.Provider,
	)

	if err := worker.Run(); err != nil {
		logx.Fatal("Email worker failed", "error", err)
	}

	logx.Info("Email worker stopped")
}
