// go:build test
package middleware

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	ctx = ContextWithUser(ctx, u)
	assert.Equal(t, u, UserFromCtx(ctx))
}

func TestContextWithUser_Nil(t *testing.T) {
	u := uuid.New()
	ctx := ContextWithUser(nil, u)
	require.NotNil(t, ctx)
}

func TestRequestWithUser(t *testing.T) {
	u := uuid.New()
	r := httptest.NewRequest("", "/", nil)
	r = RequestWithUser(r, u)
	assert.NotNil(t, r)
	assert.Equal(t, u, UserFromCtx(r.Context()))
}

func TestResponse(t *testing.T) {
	
}
