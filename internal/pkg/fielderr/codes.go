package fielderr

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

const (
	CodeBadRequest = iota
	CodeNotFound
	CodeInternal
	CodeUnauthorized
	CodeConflict
	CodeForbidden
)

var httpCodes = map[int]int{
	CodeBadRequest:   http.StatusBadRequest,
	CodeNotFound:     http.StatusNotFound,
	CodeInternal:     http.StatusInternalServerError,
	CodeUnauthorized: http.StatusUnauthorized,
	CodeForbidden:    http.StatusForbidden,
	CodeConflict:     http.StatusConflict,
}

var grpcCodes = map[int]codes.Code{
	CodeBadRequest:   codes.InvalidArgument,
	CodeInternal:     codes.Internal,
	CodeUnauthorized: codes.Unauthenticated,
	CodeConflict:     codes.InvalidArgument,
	CodeNotFound:     codes.NotFound,
	CodeForbidden:    codes.PermissionDenied,
}
