//go:generate mockgen --source=server.go --destination=mocks/service.go --package=mocks
package httpctrl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap/zapcore"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"

	_ "github.com/vlad-marlo/godo/docs"
	"github.com/vlad-marlo/godo/internal/config"
	mw "github.com/vlad-marlo/godo/internal/controller/http/middleware"
	"github.com/vlad-marlo/godo/internal/model"
)

var (
	ErrNilPointer = errors.New("nil pointer reference")
)

type Service interface {
	// Ping checks access to server.
	Ping(ctx context.Context) error
	// CreateToken create new jwt token for refresh and access to server if auth credits are correct.
	CreateToken(ctx context.Context, username, password, token string) (*model.CreateTokenResponse, error)
	// RegisterUser create record about user in storage and prepares response to user.
	RegisterUser(ctx context.Context, email, password string) (*model.User, error)
	// GetUserFromToken is helper function that decodes jwt token from t and check existing of user which id is provided
	// in token claims.
	GetUserFromToken(ctx context.Context, t string) (string, error)
	// CreateGroup create new group.
	CreateGroup(ctx context.Context, user, name, description string) (*model.CreateGroupResponse, error)
}

type Server struct {
	*chi.Mux
	cfg     *config.Config
	srv     Service
	log     *zap.Logger
	http    *http.Server
	manager *autocert.Manager
}

// New ...
func New(srv Service, cfg *config.Config, log *zap.Logger) *Server {
	s := &Server{
		Mux: chi.NewMux(),
		srv: srv,
		cfg: cfg,
		log: log,
		manager: &autocert.Manager{
			Cache:      autocert.DirCache("cache-dir"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(cfg.HTTPS.AllowedHosts...),
		},
	}
	s.configureMW()
	s.configureRoutes()
	s.http = &http.Server{
		Handler: s,
		Addr:    fmt.Sprintf("%s:%d", s.cfg.Server.Addr, s.cfg.Server.Port),
	}
	return s
}

// Stop graceful stops HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	if s == nil {
		return ErrNilPointer
	}
	return s.http.Shutdown(ctx)
}

// Start starts HTTP server.
func (s *Server) Start(context.Context) error {
	if s == nil {
		return ErrNilPointer
	}
	if s.cfg.HTTPS.CertFile == "" || s.cfg.HTTPS.KeyFile == "" || len(s.cfg.HTTPS.AllowedHosts) == 0 {
		return s.startHTTP()
	}
	return s.startHTTPS()
}

// startHTTPS ...
func (s *Server) startHTTPS() error {
	s.log.Info(
		"starting https serving",
		zap.String("https-addr", s.http.Addr),
	)
	if s.cfg.Server.Port == 0 {
		return fmt.Errorf("port is not defined")
	}

	ln, err := net.Listen("tcp", s.http.Addr)
	if err != nil {
		return fmt.Errorf("register new listener: %w", err)
	}

	go func() {
		err := s.http.ServeTLS(ln, s.cfg.HTTPS.CertFile, s.cfg.HTTPS.KeyFile)
		if !errors.Is(err, http.ErrServerClosed) {
			s.log.Fatal("unknown error while server starting", zap.Error(err))
		}
	}()

	return nil
}

// startHTTP ...
func (s *Server) startHTTP() error {
	s.log.Info(
		"starting http serving",
		zap.String("http-addr", s.http.Addr),
	)
	if s.cfg.Server.Port == 0 {
		return fmt.Errorf("port is not defined")
	}

	ln, err := net.Listen("tcp", s.http.Addr)
	if err != nil {
		return fmt.Errorf("register new listener: %w", err)
	}

	go func() {
		if err := s.http.Serve(ln); err != nil {
			s.log.Fatal("serve http", zap.Error(err))
		}
	}()
	return nil
}

// configureMW ...
func (s *Server) configureMW() {
	s.Use(
		middleware.RequestID,
		middleware.Recoverer,
		mw.LogRequest(s.log),
	)
}

// configureRoutes ...
func (s *Server) configureRoutes() {
	var prefix string
	if s.cfg.HTTPS.KeyFile == "" || s.cfg.HTTPS.CertFile == "" || len(s.cfg.HTTPS.AllowedHosts) == 0 {
		prefix = "http"
	} else {
		prefix = "https"
	}

	s.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("%s://%s:%d/docs/doc.json", prefix, s.cfg.Server.Addr, s.cfg.Server.Port)),
	))
	s.Route("/api/v1", func(r chi.Router) {
		r.HandleFunc("/ping", s.Ping)
		r.Route("/users", func(r chi.Router) {
			r.Post("/register", s.RegisterUser)
			r.Post("/token", s.CreateToken)
		})
		r.With(mw.AuthChecker(s.srv))
	})
}

// respond ...
func (s *Server) respond(w http.ResponseWriter, code int, data interface{}, fields ...zap.Field) {
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
		s.log.Log(lvl, "respond", fields...)
	}
}
