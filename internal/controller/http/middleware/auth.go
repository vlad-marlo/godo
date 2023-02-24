//go:generate mockgen --source=auth.go --destination=mocks/service.go --package=mocks
package middleware

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vlad-marlo/godo/internal/pkg/fielderr"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

// Service provide getting user from token.
type Service interface {
	GetUserFromToken(ctx context.Context, t string) (string, error)
}

const reqIDField = "request_id"

// userInCtxKey is key for context to store and get access to user data.
type userInCtxKey struct{}

// AuthChecker is mw that checks authorization header and validates request.
// Middleware adds user id to context that developer can get by userInCtxKey key
func AuthChecker(srv Service) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("authorization")
			reqID := middleware.GetReqID(r.Context())

			if !strings.Contains("Bearer ", token) && !strings.Contains("Authorization ", token) {
				w.WriteHeader(http.StatusUnauthorized)

				_, err := w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
				if err != nil {
					zap.L().Error("http: ResponseWriter: Write", zap.Error(err), zap.String(reqIDField, reqID))
				}

				return
			}

			u, err := srv.GetUserFromToken(r.Context(), token)
			if err != nil {

				if fErr, ok := err.(*fielderr.Error); ok {
					zap.L().Debug("get user from token", append(fErr.Fields(), zap.Error(err), zap.String(reqIDField, reqID))...)
					if err = json.NewEncoder(w).Encode(fErr.Data()); err != nil {
						zap.L().Error("encode data", zap.Error(err))
					}
					return
				}

				zap.L().Debug("get user from token", zap.Error(err), zap.String(reqIDField, reqID))

				w.WriteHeader(http.StatusUnauthorized)

				_, err = w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
				if err != nil {
					zap.L().Error("http: ResponseWriter: Write", zap.Error(err), zap.String(reqIDField, reqID))
				}

				return
			}

			r = r.WithContext(context.WithValue(r.Context(), userInCtxKey{}, u))
			next.ServeHTTP(w, r)
		})
	}
}

// UserFromCtx must be used with AuthChecker middleware. To get user from request you must pass *http.Request.Context() into func.
func UserFromCtx(ctx context.Context) string {
	return ctx.Value(userInCtxKey{}).(string)
}
