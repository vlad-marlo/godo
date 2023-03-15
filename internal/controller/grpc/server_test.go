package grpc

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vlad-marlo/godo/internal/config"
	"github.com/vlad-marlo/godo/internal/service/mocks"
	"go.uber.org/zap"
	"testing"
)

func TestNew(t *testing.T) {
	cfg := new(config.Config)
	cfg.Server.Port = 1234
	s := New(nil, cfg, zap.L())
	assert.Equal(t, cfg, s.cfg)
	assert.Equal(t, zap.L(), s.logger)
	assert.Nil(t, s.srv)
}

func TestStart(t *testing.T) {
	cfg := new(config.Config)
	cfg.Server.Port = 1234
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)
	s := New(srv, cfg, zap.L())
	assert.NoError(t, s.Start(context.Background()))
	assert.Error(t, s.Start(context.Background()))
	assert.NoError(t, s.Stop(context.Background()))
}
