package fielderr

import (
	"go.uber.org/zap"
	"net/http"

	"google.golang.org/grpc/codes"
)

// Error is custom error that is using in service to deliver error to controllers with prepared statuses and log fields.
type Error struct {
	// msg is error message that will be returned when Error() is called.
	msg string
	// Data stores http response if it must be returned back to user.
	Data any
	// Code must be internal code from this pkg.
	Code int
	// Fields is additional field for zap logger.
	Fields []zap.Field
}

// New creates new error with provided fields.
func New(msg string, data any, code int, fields ...zap.Field) error {
	return &Error{msg, data, code, fields}
}

// Error return error message.
func (f *Error) Error() string {
	return f.msg
}

// CodeHTTP returns http Code that is equal to custom one.
func (f *Error) CodeHTTP() int {
	if c, ok := httpCodes[f.Code]; ok {
		return c
	}
	return http.StatusInternalServerError
}

// CodeGRPC returns grpc status Code that is equal to custom one.
func (f *Error) CodeGRPC() codes.Code {
	c, ok := grpcCodes[f.Code]
	if !ok {
		return codes.Unknown
	}
	return c
}

func (f *Error) With(fields ...zap.Field) {
	f.Fields = append(f.Fields, fields...)
}
