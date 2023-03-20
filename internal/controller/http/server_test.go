package httpctrl

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vlad-marlo/godo/internal/config"
	"github.com/vlad-marlo/godo/internal/service/mocks"
	"go.uber.org/zap"
	"testing"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)
	s := New(srv, config.New(), zap.L())
	require.NotNil(t, s)
}

func TestServer_Start(t *testing.T) {
	var s *Server
	//goland:noinspection ALL
	err := s.Start(context.Background())
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNilPointer)
}

//goland:noinspection ALL
func TestServer_Stop(t *testing.T) {
	ctrl := gomock.NewController(t)

	var s *Server
	err := s.Stop(context.Background())
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNilPointer)
	srv := mocks.NewMockInterface(ctrl)
	s = New(srv, config.New(), zap.L())
	assert.NotNil(t, s)
	assert.NoError(t, s.Stop(context.Background()))
}
