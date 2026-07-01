// Package app owns the application lifecycle: it loads configuration,
// builds shared infrastructure (DB, Redis, event bus, workers), wires
// every domain service and gRPC handler, serves gRPC, and shuts down
// on SIGINT/SIGTERM. All wiring is compile-time explicit.
package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"

	"github.com/isyll/go-api-starter/internal/events"
	grpcserver "github.com/isyll/go-api-starter/internal/grpc"
	"github.com/isyll/go-api-starter/internal/infra/cache"
	database "github.com/isyll/go-api-starter/internal/infra/db"
	"github.com/isyll/go-api-starter/internal/monitor"
	"github.com/isyll/go-api-starter/internal/worker/emails"
	"github.com/isyll/go-api-starter/internal/worker/notifications"
	"github.com/isyll/go-api-starter/pkg/config"
	appenv "github.com/isyll/go-api-starter/pkg/env"
	"github.com/isyll/go-api-starter/pkg/firebase"
	"github.com/isyll/go-api-starter/pkg/idenc"
	"github.com/isyll/go-api-starter/pkg/logger"
	apptoken "github.com/isyll/go-api-starter/pkg/token"
)

// App is the application container.
type App struct {
	StartTime time.Time
	Infra     *Infrastructure

	server   *grpcserver.Server
	listener net.Listener

	bgCtx    context.Context
	bgCancel context.CancelFunc
}

// New creates an App and stamps the start time.
func New() *App {
	bgCtx, bgCancel := context.WithCancel(context.Background())
	return &App{StartTime: time.Now(), bgCtx: bgCtx, bgCancel: bgCancel}
}

// Initialize loads config and builds every shared singleton.
func (a *App) Initialize() error {
	cfgs, err := config.LoadAllConfigs()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	envName := appenv.InitApp()
	logx := logger.New(envName)
	logx.Info("initializing "+cfgs.App.Info.Name, "version", cfgs.App.Info.Version, "env", envName)

	db, err := database.InitDatabase(cfgs.Database, database.RoleApp, logx)
	if err != nil {
		return fmt.Errorf("database init: %w", err)
	}

	rdb, err := cache.InitRedis(cfgs.Redis)
	if err != nil {
		return fmt.Errorf("redis init: %w", err)
	}

	idCfg := cfgs.Security.IDObfuscation
	idEncoder := idenc.NewSqidsEncoder(idCfg.Alphabet, idCfg.MinLength)
	idenc.SetGlobalEncoder(idEncoder)

	accessTokenManager := apptoken.NewRedisAccessTokenManager(
		rdb, cfgs.Security.Auth.OAT.AccessTokenExpiry,
	)
	cacheManager := cache.NewCacheManager(rdb, cfgs.Redis.Cache.Prefix)

	fcm := a.initFCM(envName, cfgs, logx)
	d := buildDispatchers(cfgs, db, logx)

	a.Infra = &Infrastructure{
		StartTime:          a.StartTime,
		DB:                 db,
		Cache:              rdb,
		Config:             cfgs,
		Logger:             logx,
		IDEncoder:          idEncoder,
		AccessTokenManager: accessTokenManager,
		CacheManager:       cacheManager,
		FCM:                fcm,
		Notifications:      d.notif,
		Emails:             d.email,
		EventBus:           d.eventBus,
		EventBusDispatcher: d.eventAsynq,
		OutboxRepo:         d.outboxRepo,
	}
	return nil
}

// initFCM returns an FCM client, or nil when Firebase is unconfigured.
func (a *App) initFCM(env string, cfgs *config.Configs, logx *logger.Logger) *messaging.Client {
	fb, err := firebase.InitFirebase(env, cfgs, logx)
	if err != nil {
		logx.Warn("firebase disabled (push notifications off)", "error", err)
		return nil
	}
	client, err := fb.GetMessagingClient(context.Background())
	if err != nil {
		logx.Warn("fcm messaging unavailable", "error", err)
		return nil
	}
	return client
}

type dispatcherBundle struct {
	notif      notifications.Dispatcher
	email      emails.Dispatcher
	eventAsynq *events.AsynqDispatcher
	outboxRepo *events.OutboxRepository
	eventBus   *events.Bus
}

func buildDispatchers(cfgs *config.Configs, db *gorm.DB, logx *logger.Logger) dispatcherBundle {
	addr, password := cfgs.Redis.Credentials()

	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: addr, Password: password})
	eventDispatcher := events.NewAsynqDispatcher(asynqClient, logx)
	outboxRepo := events.NewOutboxRepository(db, logx)

	return dispatcherBundle{
		notif:      notifications.NewDispatcher(addr, password, cfgs.Notifications, logx),
		email:      emails.NewDispatcher(addr, password, cfgs.Email, logx),
		eventAsynq: eventDispatcher,
		outboxRepo: outboxRepo,
		eventBus:   events.NewWithOutbox(eventDispatcher, outboxRepo, logx),
	}
}

// Bootstrap wires services, subscribes events, starts background
// goroutines, and prepares the gRPC listener.
func (a *App) Bootstrap() error {
	deps := a.buildGRPCDeps()

	WireEventSubscriptions(a.Infra.EventBus, &EventHandlerDeps{
		DB:           a.Infra.DB,
		CacheManager: a.Infra.CacheManager,
		Logger:       a.Infra.Logger,
	})

	a.startBackground()

	a.server = grpcserver.New(deps)

	addr := a.Infra.Config.App.GetServerAddress()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	a.listener = lis
	return nil
}

// startBackground launches the outbox drain (opt-in), the outbox
// metrics ticker, and the dead-letter queue monitor.
func (a *App) startBackground() {
	if a.Infra.Config.Events.Outbox.DrainOnAPI {
		go a.Infra.EventBus.DrainOutbox(a.bgCtx, a.Infra.Config.Events.Outbox.Interval)
	}

	go func(ctx context.Context) {
		ticker := time.NewTicker(a.Infra.Config.Events.Outbox.MetricsInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := a.Infra.OutboxRepo.UpdateMetrics(ctx); err != nil {
					a.Infra.Logger.Warn("outbox metrics", "error", err)
				}
			}
		}
	}(a.bgCtx)

	addr, password := a.Infra.Config.Redis.Credentials()
	deadMon := monitor.NewDeadQueueMonitor(
		addr, password, 5*time.Minute,
		[]string{"high", "normal", "low", "events:dispatch"},
		a.Infra.Logger,
	)
	go deadMon.Run(a.bgCtx)
}

// Start serves gRPC in a background goroutine.
func (a *App) Start() {
	a.Infra.Logger.Info(
		"gRPC server starting",
		"address", a.listener.Addr().String(),
		"startup", time.Since(a.StartTime).String(),
	)
	go func() {
		if err := a.server.Serve(a.listener); err != nil {
			a.Infra.Logger.Fatal("gRPC serve failed", "error", err)
		}
	}()
}

// AwaitShutdown blocks until a termination signal, then drains and
// closes every resource.
func (a *App) AwaitShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	a.Infra.Logger.Info("shutdown signal received", "signal", sig.String())

	a.bgCancel()
	a.server.GracefulStop()

	if sqlDB, err := a.Infra.DB.DB(); err == nil {
		_ = sqlDB.Close()
	}
	if a.Infra.Cache != nil {
		_ = a.Infra.Cache.Close()
	}
	if a.Infra.Notifications != nil {
		_ = a.Infra.Notifications.Close()
	}
	if a.Infra.Emails != nil {
		_ = a.Infra.Emails.Close()
	}
	if a.Infra.EventBusDispatcher != nil {
		_ = a.Infra.EventBusDispatcher.Close()
	}
	a.Infra.Logger.Sync()
	a.Infra.Logger.Info("shutdown complete")
}
