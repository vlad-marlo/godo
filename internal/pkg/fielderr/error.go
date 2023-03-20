package fielderr

import (
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
	"net/http"

	"google.golang.org/grpc/codes"
)

// Error is custom error that is using in service to deliver error to controllers with prepared statuses and log fields.
type Error struct {
	// msg is error message that will be returned when Error() is called.
	msg string
	// Data stores http response if it must be returned back to user.
	data any
	// Code must be internal code from this pkg.
	code int
	// Fields is additional field for zap logger.
	fields []zap.Field
	// parent is parent error
	parent error
}

// New creates new error with provided fields.
func New(msg string, data any, code int, fields ...zap.Field) *Error {
	return &Error{msg, data, code, fields, nil}
}

// Error return error message.
func (f *Error) Error() string {
	if f == nil {
		return ""
	}
	return f.msg
}

// CodeHTTP returns http Code that is equal to custom one.
func (f *Error) CodeHTTP() int {
	if f == nil {
		return http.StatusInternalServerError
	}
	if c, ok := httpCodes[f.code]; ok {
		return c
	}
	return http.StatusInternalServerError
}

// CodeGRPC returns grpc status Code that is equal to custom one.
func (f *Error) CodeGRPC() codes.Code {
	if f == nil {
		return codes.Unknown
	}
	c, ok := grpcCodes[f.code]
	if !ok {
		return codes.Unknown
	}
	return c
}

// With create new error object that copies error fields instead of Fields
func (f *Error) With(fields ...zap.Field) *Error {
	if f == nil {
		return &Error{fields: fields}
	}
	return &Error{
		msg:    f.msg,
		data:   f.data,
		code:   f.code,
		fields: append(f.fields, fields...),
		parent: f,
	}
}

// Data return data to return to user.
func (f *Error) Data() any {
	if f == nil {
		return nil
	}
	return f.data
}

// ErrGRPC returns prepared grpc error with msg message and correct grpc code.
func (f *Error) ErrGRPC() error {
	if f == nil {
		return status.Error(codes.Unknown, "")
	}
	return status.Error(f.CodeGRPC(), f.msg)
}

// Code return internal code.
func (f *Error) Code() int {
	if f == nil {
		return 0
	}
	return f.code
}

// Fields return zap fields.
func (f *Error) Fields() []zap.Field {
	if f == nil {
		return nil
	}
	return f.fields
}

// Unwrap make available to use errors.Is with *Error.
func (f *Error) Unwrap() error {
	if f == nil {
		return nil
	}
	return f.parent
}
