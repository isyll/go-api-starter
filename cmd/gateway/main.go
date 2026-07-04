// Command gateway runs an optional HTTP/JSON reverse proxy in front of the
// gRPC API using grpc-gateway. It is opt-in and disabled by default: it is not
// part of `just run` or the default compose services. All routing is derived
// from the google.api.http annotations in the proto files, so no route is
// hand-maintained here.
package main

import (
	"context"
	"errors"
	"flag"
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

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type registrar func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error

func main() {
	grpcAddr := flag.String("grpc", envOr("GATEWAY_GRPC_ADDR", "localhost:50051"), "upstream gRPC server address")
	httpAddr := flag.String("http", envOr("GATEWAY_HTTP_ADDR", ":8080"), "HTTP listen address")
	flag.Parse()

	if err := run(*grpcAddr, *httpAddr); err != nil {
		log.Fatalf("gateway: %v", err)
	}
}

func run(grpcAddr, httpAddr string) error {
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
		if err := register(ctx, mux, grpcAddr, opts); err != nil {
			return err
		}
	}

	srv := &http.Server{Addr: httpAddr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("gateway listening on %s, proxying to %s", httpAddr, grpcAddr)
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

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
