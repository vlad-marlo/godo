package fielderr

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestNewError(t *testing.T) {
	msg, data := "err", map[string]interface{}{"xd": "ds"}
	err1 := New(msg, data, CodeInternal)
	err2 := New(msg, data, CodeInternal)
	assert.NotErrorIs(t, err1, err2)
	assert.Equal(t, err1, err2)
	wrappedErr1 := fmt.Errorf("err: %w", err1)
	assert.ErrorIs(t, wrappedErr1, err1)
}

func TestFieldError_Error(t *testing.T) {
	for i := 0; i < 1000; i++ {
		msg := uuid.NewString()
		var err error = &Error{
			data: nil,
			msg:  msg,
		}
		assert.Equal(t, err.Error(), msg)
	}
}

func TestFieldError_Fields(t *testing.T) {
	fields := map[string]interface{}{}
	var err = &Error{data: fields}
	assert.Equal(t, err.Data(), fields)
}

func TestError_CodeGRPC(t *testing.T) {
	for k, v := range grpcCodes {
		assert.Equal(t, v, (&Error{code: k}).CodeGRPC())
	}
	assert.Equal(t, codes.Unknown, (&Error{code: 123}).CodeGRPC())
}

func TestError_CodeHTTP(t *testing.T) {
	for k, v := range httpCodes {
		assert.Equal(t, v, (&Error{code: k}).CodeHTTP())
	}
	assert.Equal(t, http.StatusInternalServerError, (&Error{code: 123}).CodeHTTP())
}

func TestErrorAs(t *testing.T) {
	msg := "msg"
	data := map[string]any{
		"xd":  nil,
		"bad": 12331,
	}
	code := CodeInternal

	err := New(msg, data, code)
	newErr := err.With(zap.String("string", "sdf"))
	assert.ErrorIs(t, newErr, err)
	assert.NotEqual(t, err, newErr)
}

func TestError_Fields(t *testing.T) {
	fields := []zap.Field{
		zap.String("xd", "xd"),
	}
	err := &Error{fields: fields}
	assert.Equal(t, err.Fields(), err.fields)
}
