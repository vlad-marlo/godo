package httpctrl

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/vlad-marlo/godo/internal/config"
	"github.com/vlad-marlo/godo/internal/model"
	"go.uber.org/zap"
)

var (
	TestUser1 = &model.User{
		ID:    uuid.New(),
		Pass:  "test password",
		Email: "example@ex.com",
	}
	TestTokenRequest = &model.CreateTokenRequest{
		Email:     TestUser1.Email,
		Password:  TestUser1.Pass,
		TokenType: "auth",
	}
)

// TestServer is helper function that creates http server
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

// reqWithData add key and val to chi url params.
//
// key - chi url param name
// val - chi url param value
func reqWithData(t testing.TB, r *http.Request, key, val string) *http.Request {
	t.Helper()
	rCtx := chi.NewRouteContext()
	rCtx.URLParams.Add(key, val)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rCtx))

	require.Equal(t, val, chi.URLParam(r, key))

	return r
}

// reqWithGroup adds chi url param to context.
func reqWithGroup(t testing.TB, r *http.Request, val string) *http.Request {
	return reqWithData(t, r, "group_id", val)
}

// reqWithTask is helper func to call reqWithData with task_id field.
func reqWithTask(t testing.TB, r *http.Request, val string) *http.Request {
	return reqWithData(t, r, "task_id", val)
}
