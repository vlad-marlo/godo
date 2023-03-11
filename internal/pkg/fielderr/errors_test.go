package fielderr

import (
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestNewError(t *testing.T) {
	msg, data := "err", map[string]any{"xd": "ds"}
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
	var err *Error
	assert.Equal(t, "", err.Error())
}

func TestFieldError_Fields(t *testing.T) {
	fields := map[string]any{}
	var err = &Error{data: fields}
	assert.Equal(t, fields, err.Data())
	assert.Equal(t, nil, (*Error)(nil).Data())
}

func TestError_CodeGRPC(t *testing.T) {
	for k, v := range grpcCodes {
		assert.Equal(t, v, (&Error{code: k}).CodeGRPC())
	}
	assert.Equal(t, codes.Unknown, (&Error{code: 123}).CodeGRPC())
	assert.Equal(t, codes.Unknown, (*Error)(nil).CodeGRPC())
}

func TestError_CodeHTTP(t *testing.T) {
	for k, v := range httpCodes {
		assert.Equal(t, v, (&Error{code: k}).CodeHTTP())
	}
	assert.Equal(t, http.StatusInternalServerError, (&Error{code: 123}).CodeHTTP())
	assert.Equal(t, http.StatusInternalServerError, (*Error)(nil).CodeHTTP())
}

func TestErrorIs(t *testing.T) {
	msg := "msg"
	data := map[string]any{
		"xd":  nil,
		"bad": 12331,
	}
	code := CodeInternal

	err := New(msg, data, code)
	newErr := err.With(zap.String("string", "sdf"), zap.Error(nil))
	assert.ErrorIs(t, error(newErr), error(err))
	assert.NotEqual(t, error(err), error(newErr))
	assert.Equal(t, (error)(nil), (*Error)(nil).Unwrap())
}

func TestError_Fields(t *testing.T) {
	fields := []zap.Field{
		zap.String("xd", "xd"),
	}
	err := &Error{fields: fields}
	assert.Equal(t, err.Fields(), err.fields)
	err = nil
	assert.Equal(t, ([]zap.Field)(nil), err.Fields())
}

func TestError_Err(t *testing.T) {
	err := (*Error)(nil)
	assert.Equal(t, status.Error(codes.Unknown, ""), err.ErrGRPC())
	err = &Error{
		msg:  "some message",
		code: CodeForbidden,
	}
	assert.Equal(t, status.Error(grpcCodes[CodeForbidden], err.msg), err.ErrGRPC())
}

func TestError_With(t *testing.T) {
	fields := []zap.Field{zap.String("", "")}
	err := (*Error)(nil).With(fields...)
	assert.Equal(t, &Error{fields: fields}, err)
}

func TestError_Code(t *testing.T) {
	assert.Equal(t, 0, (*Error)(nil).Code())
}
