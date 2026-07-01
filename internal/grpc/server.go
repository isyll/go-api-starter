package grpc

import (
	"net"

	"github.com/isyll/go-api-starter/internal/domain/auth"
	apiv1 "github.com/isyll/go-api-starter/internal/gen/api/v1"
	"github.com/isyll/go-api-starter/pkg/config"
	"github.com/isyll/go-api-starter/pkg/logger"
	apptoken "github.com/isyll/go-api-starter/pkg/token"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Deps is everything the gRPC server needs to serve requests.
type Deps struct {
	Logger   *logger.Logger
	Config   *config.Configs
	Tokens   apptoken.AccessTokenManager
	Sessions auth.DeviceSessionRepository
	Auth     *AuthServer
	User     *UserServer
	Admin    *AdminServer
	Health   *HealthServer
}

// Server wraps a configured *grpc.Server.
type Server struct {
	grpc   *grpc.Server
	logger *logger.Logger
}

// New builds the gRPC server, wires the interceptor chain, and
// registers every service plus reflection and the standard health
// service.
func New(d Deps) *Server {
	ic := &interceptors{
		tokens:   d.Tokens,
		sessions: d.Sessions,
		cfg:      d.Config,
		logger:   d.Logger,
	}

	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(
		ic.recoveryUnary,
		ic.requestIDUnary,
		ic.loggingUnary,
		ic.authUnary,
	))

	apiv1.RegisterAuthServiceServer(srv, d.Auth)
	apiv1.RegisterUserServiceServer(srv, d.User)
	apiv1.RegisterAdminServiceServer(srv, d.Admin)
	apiv1.RegisterHealthServiceServer(srv, d.Health)

	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(srv, hs)

	reflection.Register(srv)

	return &Server{grpc: srv, logger: d.Logger}
}

// Serve blocks serving on the listener until GracefulStop is called.
func (s *Server) Serve(lis net.Listener) error {
	return s.grpc.Serve(lis)
}

// GracefulStop drains in-flight RPCs and stops the server.
func (s *Server) GracefulStop() {
	s.grpc.GracefulStop()
}
