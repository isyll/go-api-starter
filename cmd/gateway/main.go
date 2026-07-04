// Command gateway runs an optional HTTP/JSON reverse proxy in front of the
// gRPC API using grpc-gateway. It is opt-in and off by default: it exits unless
// gateway.enabled is set, and it is never started by `just run` or the default
// compose services. All routing is derived from the google.api.http annotations
// in the proto files, so no route is hand-maintained here.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	adminv1 "github.com/isyll/go-grpc-starter/gen/admin/v1"
	authv1 "github.com/isyll/go-grpc-starter/gen/auth/v1"
	healthv1 "github.com/isyll/go-grpc-starter/gen/health/v1"
	userv1 "github.com/isyll/go-grpc-starter/gen/user/v1"
	"github.com/isyll/go-grpc-starter/pkg/config"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type registrar func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error

func main() {
	config.LoadEnvFile()

	gw, err := config.LoadConfig[config.GatewayConfig]("configs/gateway.yaml")
	if err != nil {
		log.Fatalf("gateway: load config: %v", err)
	}
	if !gw.Enabled {
		log.Print("gateway disabled (set gateway.enabled to run)")
		return
	}

	app, err := config.LoadConfig[config.AppConfig]("configs/api.yaml")
	if err != nil {
		log.Fatalf("gateway: load server config: %v", err)
	}
	upstream := gw.UpstreamAddr
	if upstream == "" {
		upstream = "localhost:" + app.Server.Port
	}

	if err := run(gw.ListenAddr, upstream); err != nil {
		log.Fatalf("gateway: %v", err)
	}
}

func run(listenAddr, upstreamAddr string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(headerMatcher))
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	for _, register := range []registrar{
		authv1.RegisterAuthServiceHandlerFromEndpoint,
		userv1.RegisterUserServiceHandlerFromEndpoint,
		adminv1.RegisterAdminServiceHandlerFromEndpoint,
		healthv1.RegisterHealthServiceHandlerFromEndpoint,
	} {
		if err := register(ctx, mux, upstreamAddr, opts); err != nil {
			return err
		}
	}

	srv := &http.Server{Addr: listenAddr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("gateway listening on %s, proxying to %s", listenAddr, upstreamAddr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// headerMatcher forwards the headers the interceptors expect (bearer token,
// language, request id) verbatim as gRPC metadata; grpc-gateway would
// otherwise rename them under a grpcgateway- prefix.
func headerMatcher(key string) (string, bool) {
	switch strings.ToLower(key) {
	case "authorization", "accept-language", "lang", "x-request-id":
		return strings.ToLower(key), true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}
