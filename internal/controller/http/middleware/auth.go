//go:generate mockgen --source=auth.go --destination=mocks/service.go --package=mocks
package middleware

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/vlad-marlo/godo/internal/pkg/fielderr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
)

// Service provide getting user from token.
type Service interface {
	// GetUserFromToken return id of user, who claim provided token.
	GetUserFromToken(ctx context.Context, t string) (uuid.UUID, error)
}

const reqIDField = "request_id"

// userInCtxKey is key for context to store and get access to user data.
type userInCtxKey struct{}

// AuthChecker is mw that checks authorization header and validates request.
// Middleware adds user id to context that developer can get by userInCtxKey key
func AuthChecker(srv Service) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			zap.L().Info("header", zap.Any("header", r.Header))
			token := r.Header.Get("authorization")
			reqID := middleware.GetReqID(r.Context())

			u, err := srv.GetUserFromToken(r.Context(), token)
			if err != nil {

				if fErr, ok := err.(*fielderr.Error); ok {
					respond(w, fErr.CodeHTTP(), fErr.Data(), append(fErr.Fields(), zap.Error(err), zap.String(reqIDField, reqID))...)
					return
				}

				respond(w, http.StatusInternalServerError, nil, zap.Error(err), zap.String(reqIDField, reqID))
				return
			}

			next.ServeHTTP(w, WithUser(r, u))
		})
	}
}

// respond helper function to write response to user.
func respond(w http.ResponseWriter, code int, data interface{}, fields ...zap.Field) {
	var lvl zapcore.Level
	switch {
	case code >= 500:
		lvl = zap.DPanicLevel
	case code >= 400:
		lvl = zap.ErrorLevel
	default:
		lvl = zap.DebugLevel
	}

	w.Header().Set("content-type", "application/json")

	if data == nil {
		data = http.StatusText(code)
	}

	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fields = append(fields, zap.Error(err))
	}

	if len(fields) > 0 {
		zap.L().Log(lvl, "respond", fields...)
	}
}

// WithUser adds user to request's context.
// To get user id from request use UserFromContext.
func WithUser(r *http.Request, u uuid.UUID) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), userInCtxKey{}, u))
}

// UserFromCtx must be used with AuthChecker middleware. To get user from request you must pass *http.Request.Context() into func.
func UserFromCtx(ctx context.Context) uuid.UUID {
	return ctx.Value(userInCtxKey{}).(uuid.UUID)
}
