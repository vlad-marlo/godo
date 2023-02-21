package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/vlad-marlo/godo/internal/controller/grpc/mocks"
	"github.com/vlad-marlo/godo/pkg/proto/api/v1/pb"
)

var errUnknown = errors.New("")

func TestServer_Ping_Positive(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockService(ctrl)
	srv.EXPECT().Ping(gomock.Any()).Return(nil).AnyTimes()
	s := TestServer(t, srv)
	_, err := s.Ping(context.Background(), &pb.PingRequest{})
	require.NoError(t, err)
}

func TestServer_Ping_Negative(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockService(ctrl)
	srv.EXPECT().Ping(gomock.Any()).Return(errUnknown)
	s := &Server{
		UnimplementedGodoServer: pb.UnimplementedGodoServer{},
		srv:                     srv,
		logger:                  zap.L(),
	}
	_, err := s.Ping(context.Background(), nil)
	require.Error(t, err)
}
