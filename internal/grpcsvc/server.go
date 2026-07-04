package grpcsvc

import (
	"net"

	adminv1 "github.com/isyll/go-grpc-starter/gen/admin/v1"
	authv1 "github.com/isyll/go-grpc-starter/gen/auth/v1"
	healthv1 "github.com/isyll/go-grpc-starter/gen/health/v1"
	userv1 "github.com/isyll/go-grpc-starter/gen/user/v1"
	"github.com/isyll/go-grpc-starter/internal/auth"
	"github.com/isyll/go-grpc-starter/internal/interceptor"
	"github.com/isyll/go-grpc-starter/pkg/config"
	"github.com/isyll/go-grpc-starter/pkg/locale"
	"github.com/isyll/go-grpc-starter/pkg/logger"
	apptoken "github.com/isyll/go-grpc-starter/pkg/token"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Deps struct {
	Logger   *logger.Logger
	Config   *config.Configs
	Tokens   apptoken.AccessTokenManager
	Sessions auth.DeviceSessionRepository
	Locale   *locale.Bundle
	Auth     *AuthServer
	User     *UserServer
	Admin    *AdminServer
	Health   *HealthServer
}

type Server struct {
	grpc   *grpc.Server
	logger *logger.Logger
}

func New(d Deps) *Server {
	ic := interceptor.New(interceptor.Config{
		Tokens:   d.Tokens,
		Sessions: d.Sessions,
		Cfg:      d.Config,
		Logger:   d.Logger,
		Locale:   d.Locale,
	})

	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(ic.Unary()...))

	authv1.RegisterAuthServiceServer(srv, d.Auth)
	userv1.RegisterUserServiceServer(srv, d.User)
	adminv1.RegisterAdminServiceServer(srv, d.Admin)
	healthv1.RegisterHealthServiceServer(srv, d.Health)

	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(srv, hs)

	reflection.Register(srv)

	return &Server{grpc: srv, logger: d.Logger}
}

func (s *Server) Serve(lis net.Listener) error {
	return s.grpc.Serve(lis)
}

func (s *Server) GracefulStop() {
	s.grpc.GracefulStop()
}
