package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/vlad-marlo/godo/internal/config"
)

// New sets logger with new instance.
func New(cfg *config.Config) *zap.Logger {
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})

	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)
	var jsonEncoder zapcore.Encoder
	switch {
	case cfg.Test.Enable:
		return zap.L()
	case cfg.Server.IsDev:
		jsonEncoder = zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig())
	default:
		jsonEncoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	}

	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, consoleErrors, highPriority),
		zapcore.NewCore(jsonEncoder, consoleDebugging, lowPriority),
	)
	l := zap.
		New(core).
		With(
			zap.String("postgres db", cfg.Postgres.Name),
			zap.String("server", cfg.Server.Type),
			zap.String("addr", cfg.Server.Addr),
			zap.Uint("port", cfg.Server.Port),
		)
	zap.ReplaceGlobals(l)
	return l
}
