package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/pkg/fielderr"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/vlad-marlo/godo/internal/config"
	"github.com/vlad-marlo/godo/pkg/proto/api/v1/pb"
)

// Service ...
type Service interface {
	// Ping checks access to server.
	Ping(ctx context.Context) error
	// CreateToken create new jwt token for refresh and access to server if auth credits are correct.
	CreateToken(ctx context.Context, email, password, token string) (*model.CreateTokenResponse, error)
	// RegisterUser create record about user in storage and prepares response to user.
	RegisterUser(ctx context.Context, email, password string) (*model.User, error)
	// GetUserFromToken is helper function that decodes jwt token from t and check existing of user which id is provided
	// in token claims.
	GetUserFromToken(ctx context.Context, t string) (uuid.UUID, error)
	// CreateGroup create new group.
	CreateGroup(ctx context.Context, user uuid.UUID, name, description string) (*model.CreateGroupResponse, error)
	// CreateInvite creates invite link.
	CreateInvite(ctx context.Context, user, group uuid.UUID, role *model.Role, limit int) (*model.CreateInviteResponse, error)
	// UseInvite add user to group if invite is ok.
	UseInvite(ctx context.Context, user, group, invite uuid.UUID) error
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
		if err = s.server.Serve(ln); err != nil {
			s.logger.Warn("serve grpc", zap.Error(err))
		}
	}()
	s.logger.Info("starting GRPC server")
	return nil
}

// Stop stops GRPC server.
func (s *Server) Stop(context.Context) error {
	s.server.GracefulStop()
	return nil
}

// handleErr
func (s *Server) handleErr(msg string, err error) error {
	var fErr *fielderr.Error
	if errors.As(err, &fErr) {
		return fErr.ErrGRPC()
	}

	return s.internal(msg, err)
}
