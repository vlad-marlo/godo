package httpctrl

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/vlad-marlo/godo/internal/config"
	"github.com/vlad-marlo/godo/internal/model"
	"go.uber.org/zap"
	"testing"
)

var (
	TestUser1 = &model.User{
		ID:      uuid.New(),
		Name:    "test user 1",
		Pass:    "test password",
		IsAdmin: false,
	}
)

func TestServer(t testing.TB, srv Service) *Server {
	t.Helper()
	return &Server{
		Mux:     chi.NewMux(),
		cfg:     config.New(),
		srv:     srv,
		log:     zap.L(),
		http:    nil,
		manager: nil,
	}
}
