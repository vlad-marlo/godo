package fielderr

import (
	"fmt"
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
			Data: nil,
			msg:  msg,
		}
		assert.Equal(t, err.Error(), msg)
	}
}

func TestFieldError_Fields(t *testing.T) {
	fields := map[string]interface{}{}
	var err = &Error{Data: fields}
	assert.Equal(t, err.Data, fields)
}

func TestError_CodeGRPC(t *testing.T) {
	for k, v := range grpcCodes {
		assert.Equal(t, v, (&Error{Code: k}).CodeGRPC())
	}
	assert.Equal(t, codes.Unknown, (&Error{Code: 123}).CodeGRPC())
}

func TestError_CodeHTTP(t *testing.T) {
	for k, v := range httpCodes {
		assert.Equal(t, v, (&Error{Code: k}).CodeHTTP())
	}
	assert.Equal(t, http.StatusInternalServerError, (&Error{Code: 123}).CodeHTTP())
}
