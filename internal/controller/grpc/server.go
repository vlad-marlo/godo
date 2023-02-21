//go:generate mockgen --source=server.go --destination=mocks/service.go --package=mocks
package grpc

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/vlad-marlo/godo/internal/config"
	"github.com/vlad-marlo/godo/pkg/proto/api/v1/pb"
)

// Service ...
type Service interface {
	Ping(ctx context.Context) error
}

// Server ...
type Server struct {
	pb.UnimplementedGodoServer
	srv    Service
	cfg    *config.Config
	logger *zap.Logger
	server *grpc.Server
}

// New return new server with provided Service.
func New(srv Service, cfg *config.Config, log *zap.Logger) *Server {
	s := &Server{
		UnimplementedGodoServer: pb.UnimplementedGodoServer{},
		logger:                  log,
		srv:                     srv,
		cfg:                     cfg,
		server:                  grpc.NewServer(),
	}
	pb.RegisterGodoServer(s.server, s)
	return s
}

// Start starts GRPC server.
func (s *Server) Start(context.Context) error {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.cfg.Server.Addr, s.cfg.Server.Port))
	if err != nil {
		return fmt.Errorf("net: listen: %w", err)
	}
	go func() {
		if err := s.server.Serve(ln); err != nil {
			s.logger.Fatal("serve grpc", zap.Error(err))
		}
	}()
	s.logger.Info("starting GRPC server")
	return nil
}

func (s *Server) Stop(context.Context) error {
	s.server.GracefulStop()
	return nil
}
