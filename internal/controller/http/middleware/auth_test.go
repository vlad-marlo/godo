// go:build test
package middleware

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserFromCtx_NilContext(t *testing.T) {
	assert.Equal(t, uuid.Nil, UserFromCtx((context.Context)(nil)))
}

func TestUserFromCtx_NoUserInIt(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, uuid.Nil, UserFromCtx(ctx))
}

func TestUserFromCtx_Ok(t *testing.T) {
	ctx := context.Background()
	u := uuid.New()
	ctx = contextWithUser(ctx, u)
	assert.Equal(t, u, UserFromCtx(ctx))
}

func TestContextWithUser_Nil(t *testing.T) {
	u := uuid.New()
	ctx := contextWithUser(nil, u)
	require.NotNil(t, ctx)
}

func TestRequestWithUser(t *testing.T) {
	u := uuid.New()
	r := httptest.NewRequest("", "/", nil)
	r = RequestWithUser(r, u)
	assert.NotNil(t, r)
	assert.Equal(t, u, UserFromCtx(r.Context()))
}

func TestRespond_NilData(t *testing.T) {
	w := httptest.NewRecorder()
	respond(w, http.StatusOK, nil)
	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	data, err := json.Marshal(http.StatusText(http.StatusOK))
	require.NoError(t, err)
	assert.JSONEq(t, string(data), w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRespond_NotNilData(t *testing.T) {
	w := httptest.NewRecorder()
	data := &struct {
		Name string `json:"name"`
		Kk   uint   `json:"value"`
	}{"xd", 2}
	respond(w, http.StatusNotFound, data)
	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	got, err := json.Marshal(data)
	require.NoError(t, err)
	assert.JSONEq(t, string(got), w.Body.String())
}

func TestRespond_NotNilFields(t *testing.T) {
	w := httptest.NewRecorder()
	data := &struct {
		Name string `json:"name"`
		Kk   uint   `json:"value"`
	}{"xd", 2}
	respond(w, http.StatusNotFound, data, zap.String("sd", "sd"))
	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	got, err := json.Marshal(data)
	require.NoError(t, err)
	t.Logf("%s", got)
	assert.JSONEq(t, string(got), w.Body.String())
}
