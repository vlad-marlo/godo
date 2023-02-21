package logger

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/vlad-marlo/godo/internal/config"
)

func TestNewChangesGlobalLogger(t *testing.T) {
	globalBefore := zap.L()
	_ = os.Setenv("TEST", "false")
	l := New(config.New())
	globalAfter := zap.L()
	assert.Equal(t, globalAfter, l)
	assert.NotEqual(t, globalAfter, globalBefore)
	assert.NotEqual(t, globalBefore, l)
}

func TestNewReturnDefaultLoggerForTests(t *testing.T) {
	oldGlobal := zap.L()
	cfg := &config.Config{Test: config.Test{Enable: true}}
	l := New(cfg)
	newGlobal := zap.L()
	assert.Equal(t, oldGlobal, l)
	assert.Equal(t, oldGlobal, newGlobal)
}

func TestNewDev(t *testing.T) {
	oldGlobal := zap.L()
	cfg := &config.Config{Server: config.Server{IsDev: true}}
	l := New(cfg)
	newGlobal := zap.L()
	assert.Equal(t, newGlobal, l)
	assert.NotEqual(t, oldGlobal, newGlobal)
}

func TestNewProd(t *testing.T) {
	oldGlobal := zap.L()
	cfg := &config.Config{Server: config.Server{IsDev: false}}
	l := New(cfg)
	newGlobal := zap.L()
	assert.Equal(t, newGlobal, l)
	assert.NotEqual(t, oldGlobal, newGlobal)
}
