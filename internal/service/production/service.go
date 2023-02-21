package production

import (
	"context"
	"github.com/vlad-marlo/godo/internal/config"
	"github.com/vlad-marlo/godo/internal/pkg/fielderr"
	"github.com/vlad-marlo/godo/internal/store"
	"go.uber.org/zap"
	"net/http"
)

// Service ...
type Service struct {
	store store.Store
	cfg   *config.Config
	log   *zap.Logger
}

// New ...
func New(store store.Store, cfg *config.Config, log *zap.Logger) *Service {
	return &Service{
		store: store,
		cfg:   cfg,
		log:   log,
	}
}

// Ping ...
func (s *Service) Ping(ctx context.Context) error {
	if err := s.store.Ping(ctx); err != nil {
		return fielderr.New(
			http.StatusText(http.StatusInternalServerError),
			nil,
			fielderr.CodeInternal,
		)
	}
	return nil
}
