package grpc

import (
	"net"

	"github.com/isyll/go-grpc-starter/internal/auth"
	apiv1 "github.com/isyll/go-grpc-starter/internal/gen/api/v1"
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
	ic := &interceptors{
		tokens:        d.Tokens,
		sessions:      d.Sessions,
		cfg:           d.Config,
		logger:        d.Logger,
		localeDefLang: "en",
	}
	if d.Locale != nil {
		ic.locale = d.Locale
		ic.localeDefLang = d.Locale.DefaultLanguage()
	}

	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(
		ic.recoveryUnary,
		ic.loggingUnary,
		ic.localeUnary,
		ic.errorUnary,
		ic.requestIDUnary,
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

func (s *Server) Serve(lis net.Listener) error {
	return s.grpc.Serve(lis)
}

func (s *Server) GracefulStop() {
	s.grpc.GracefulStop()
}
