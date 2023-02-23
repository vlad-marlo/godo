package main

import (
	"context"
	"errors"
	"github.com/vlad-marlo/godo/internal/config"
	"github.com/vlad-marlo/godo/internal/controller/grpc"
	httpctrl "github.com/vlad-marlo/godo/internal/controller/http"
	"github.com/vlad-marlo/godo/internal/pkg/client/postgres"
	"github.com/vlad-marlo/godo/internal/pkg/logger"
	"github.com/vlad-marlo/godo/internal/service"
	"github.com/vlad-marlo/godo/internal/service/production"
	"github.com/vlad-marlo/godo/internal/store"
	"github.com/vlad-marlo/godo/internal/store/pgx"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	fx.New(CreateApp()).Run()
}

// CreateApp returns fx options to create HTTP application.
func CreateApp() fx.Option {
	return fx.Options(
		fx.Provide(
			logger.New,
			grpc.New,
			config.New,
			fx.Annotate(
				postgres.New,
				fx.As(new(pgx.Client)),
			),
			fx.Annotate(
				ServiceFactory,
				fx.As(new(httpctrl.Service)),
				fx.As(new(grpc.Service)),
			),
			fx.Annotate(
				pgx.New,
				fx.As(new(store.Store)),
			),
			pgx.NewGroupRepository,
			pgx.NewUserRepository,
			httpctrl.New,
		),
		//fx.WithLogger(ZapEventLogger),
		fx.Invoke(
			ValidateConfig,
			StartHTTPServer,
			CheckClient,
			StartGRPCServer,
			LoggerSyncer,
		),
	)
}

// ZapEventLogger return new event logger for fx application.
func ZapEventLogger(logger *zap.Logger) fxevent.Logger {
	return &fxevent.ZapLogger{Logger: logger}
}

// StartHTTPServer is starting http server if must.
func StartHTTPServer(lc fx.Lifecycle, h *httpctrl.Server, cfg *config.Config) {
	if !cfg.Server.EnableHTTP {
		return
	}
	lc.Append(fx.Hook{
		OnStart: h.Start,
		OnStop:  h.Stop,
	})
}

// StartGRPCServer is starting grpc server if must.
func StartGRPCServer(lc fx.Lifecycle, h *grpc.Server, cfg *config.Config) {
	if !cfg.Server.EnableGRPC {
		return
	}
	lc.Append(fx.Hook{
		OnStart: h.Start,
		OnStop:  h.Stop,
	})
}

// CheckClient checks connection to postgres server and add closing pool hook.
func CheckClient(lc fx.Lifecycle, client pgx.Client) error {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			client.P().Close()
			return nil
		},
	})
	return client.P().Ping(context.Background())
}

// ValidateConfig checks if config valid and if not logs recommendations to configure application.
func ValidateConfig(cfg *config.Config, log *zap.Logger) error {
	if ok, err := cfg.Valid(); err != nil || !ok {
		log.Error("config is not valid", zap.Bool("ok", ok), zap.Error(err))
		return errors.New("bad config")
	}
	return nil
}

// ServiceFactory return right service for server. If server is running on development mode than factory will return
// development service instead of production.
func ServiceFactory(store store.Store, cfg *config.Config, log *zap.Logger) service.Interface {
	//if cfg.Server.IsDev {
	// create development server if necessary.
	//}
	return production.New(store, cfg, log)
}

func LoggerSyncer(lc fx.Lifecycle, log *zap.Logger) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return log.Sync()
		},
	})
}