//go:generate mockgen --source=auth.go --destination=mocks/service.go --package=mocks
package middleware

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type Service interface {
	GetUserFromToken(ctx context.Context, t string) (string, error)
}

// userInCtxKey is key for context to store and get access to user data.
type userInCtxKey struct{}

// AuthChecker is mw that checks authorization header and validates request.
// Middleware adds user id to context that developer can get by userInCtxKey key
func AuthChecker(srv Service) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("authorization")
			reqID := middleware.GetReqID(r.Context())

			if !strings.Contains("Bearer ", token) {
				w.WriteHeader(http.StatusUnauthorized)

				_, err := w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
				if err != nil {
					zap.L().Error("http: ResponseWriter: Write", zap.Error(err), zap.String("request_id", reqID))
				}
				return
			}

			token = strings.TrimPrefix("Bearer ", token)

			u, err := srv.GetUserFromToken(r.Context(), token)
			if err != nil {
				zap.L().Debug("get user from token", zap.Error(err), zap.String("request_id", reqID))

				w.WriteHeader(http.StatusUnauthorized)
				_, err = w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
				if err != nil {
					zap.L().Error("http: ResponseWriter: Write", zap.Error(err), zap.String("request_id", reqID))
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
