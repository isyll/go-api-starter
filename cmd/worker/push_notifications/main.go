// Command push_notifications is the standalone Asynq worker
// binary that drains the notifications:* queues and delivers
// in-app push notifications via Firebase Cloud Messaging.
//
// It wires the Firebase admin SDK, the FCM token repository,
// and the notifications worker processor, then hands control
// to asynq.Server. Panics inside ProcessTask are caught by the
// worker's deferred recover() so a single malformed payload
// does not kill the goroutine.
package main

import (
	"context"
	"fmt"
	"log"

	database "github.com/isyll/go-api-starter/internal/infra/db"
	notifWorker "github.com/isyll/go-api-starter/internal/worker/notifications"
	"github.com/isyll/go-api-starter/pkg/config"
	appenv "github.com/isyll/go-api-starter/pkg/env"
	"github.com/isyll/go-api-starter/pkg/firebase"
	"github.com/isyll/go-api-starter/pkg/logger"
)

// main loads backend configs, opens the FCM and DB connections
// needed by the push worker, then blocks on asynq.Server until
// SIGINT / SIGTERM.
func main() {
	env := appenv.InitApp()

	cfg, err := config.LoadAllConfigs()
	if err != nil {
		log.Fatal("failed to load configurations:", err)
	}

	logx := logger.New(env)

	logx.Info("Starting push notification worker", "env", env)

	db, err := database.InitDatabase(
		cfg.Database, database.RoleAdmin, logx,
	)
	if err != nil {
		logx.Fatal("Failed to initialize database", "error", err)
	}

	firebaseClient, err := firebase.InitFirebase(env, cfg, logx)
	if err != nil {
		logx.Fatal("Failed to initialize Firebase client", "error", err)
	}

	fcmClient, err := firebaseClient.GetMessagingClient(
		context.Background(),
	)
	if err != nil {
		logx.Fatal("Failed to get FCM messaging client", "error", err)
	}

	fcmTokenRepo := notifWorker.NewFCMTokenRepository(db)
	preferencesRepo := notifWorker.NewNotificationPreferencesRepository(
		db,
	)
	templateRepo := notifWorker.NewTemplateRepository(db)
	logRepo := notifWorker.NewLogRepository(db)

	redisAddr := fmt.Sprintf(
		"%s:%d",
		cfg.Redis.Connection.Host,
		cfg.Redis.Connection.Port,
	)
	redisPassword := cfg.Redis.Connection.Password

	processor := notifWorker.NewProcessor(
		fcmClient,
		fcmTokenRepo,
		preferencesRepo,
		templateRepo,
		logRepo,
		cfg.Notifications,
		logx,
	)

	worker := notifWorker.NewWorker(
		redisAddr,
		redisPassword,
		processor,
		cfg.Notifications,
		logx,
	)

	logx.Info("Push notification worker starting",
		"redis", redisAddr,
		"concurrency", cfg.Notifications.Worker.Concurrency,
	)

	if err := worker.Run(); err != nil {
		logx.Fatal("Worker failed", "error", err)
	}

	logx.Info("Push notification worker stopped")
}
