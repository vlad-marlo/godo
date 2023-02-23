package httpctrl

import (
	"bufio"
	"encoding/json"
	"github.com/go-chi/chi/v5/middleware"
	mw "github.com/vlad-marlo/godo/internal/controller/http/middleware"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/pkg/fielderr"
	"go.uber.org/zap"
	"net/http"
)

const ZapRequestIDFieldName = "request_id"

// ReqIDField return named zap field with reqID in it.
func ReqIDField(reqID string) zap.Field {
	return zap.String(ZapRequestIDFieldName, reqID)
}

// RegisterUser creates user with provided data.
// @Tags UserCreate
// @Summary Создание пользователя
// @ID user_create
// @Accept json
// @Produce json
// @Success 201 {object} model.CreateUserResponse
// @Failure 400 {string} "Bad Request"
// @Failure 409 {string} "Conflict"
// @Failure 500 {string} "Internal Server Error"
// @Router /api/v1/users/register [post]
func (s *Server) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterUserRequest
	reqID := middleware.GetReqID(r.Context())

	buf := bufio.NewReader(r.Body)
	if err := r.Body.Close(); err != nil {
		s.respond(w, http.StatusInternalServerError, nil, ReqIDField(reqID), zap.Error(err))
		return
	}

	if err := json.NewDecoder(buf).Decode(&req); err != nil {
		s.respond(w, http.StatusBadRequest, nil, zap.Error(err), ReqIDField(reqID))
		return
	}

	u, err := s.srv.RegisterUser(r.Context(), req.Email, req.Password)
	if err != nil {
		if fieldErr, ok := err.(*fielderr.Error); ok {
			s.respond(w, fieldErr.CodeHTTP(), fieldErr.Data(), ReqIDField(reqID))
			return
		}

		s.respond(w, http.StatusInternalServerError, nil, zap.Error(err), ReqIDField(reqID))
		return
	}

	s.respond(w, http.StatusCreated, u)
}

// LoginJWT creates JWT bearer token with provided data.
// @Tags JWTCreate
// @Summary Создание пользователя
// @ID login_jwt
// @Accept json
// @Produce json
// @Success 201 {object} model.CreateJWTResponse
// @Failure 400 {string} "Bad Request"
// @Failure 401 {string} "Unauthorized"
// @Failure 500 {string} "Internal Server Error"
// @Router /api/v1/users/login/jwt [post]
func (s *Server) LoginJWT(w http.ResponseWriter, r *http.Request) {
	var req model.LoginUserRequest
	defer func() {
		_ = r.Body.Close()
	}()

	reqID := middleware.GetReqID(r.Context())

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respond(w, http.StatusBadRequest, nil, zap.Error(err), ReqIDField(reqID))
		return
	}

	u, err := s.srv.LoginUserJWT(r.Context(), req.Email, req.Password)
	if err != nil {
		if fdErr, ok := err.(*fielderr.Error); ok {
			s.respond(w, fdErr.CodeHTTP(), fdErr.Data(), ReqIDField(reqID), zap.Error(fdErr))
			return
		}

		s.respond(w, http.StatusInternalServerError, nil, zap.Error(err), ReqIDField(reqID))
		return
	}
	s.respond(w, http.StatusOK, u)
}

// Ping godoc.
// @Tags Ping
// @Summary Запрос состояния сервиса
// @ID ping
// @Accept plain
// @Produce plain
// @Success 200 {string} string "OK"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/v1/ping [get]
func (s *Server) Ping(w http.ResponseWriter, r *http.Request) {
	if err := s.srv.Ping(r.Context()); err != nil {

		if fErr, ok := err.(*fielderr.Error); ok {
			s.respond(w, fErr.CodeHTTP(), fErr.Data(), zap.Error(err))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) CreateGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	user := mw.UserFromCtx(r.Context())

	var req model.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respond(w, http.StatusBadRequest, nil, ReqIDField(reqID), zap.Error(err))
	}

	resp, err := s.srv.CreateGroup(ctx, user, req.Name, req.Description)
	if err != nil {
		fErr, ok := err.(*fielderr.Error)
		if !ok {
			s.respond(w, http.StatusBadRequest, nil, zap.Error(err), ReqIDField(reqID))
			return
		}
		s.respond(w, fErr.CodeHTTP(), fErr.Data(), append(fErr.Fields(), ReqIDField(reqID))...)
		return
	}
	s.respond(w, http.StatusCreated, resp, ReqIDField(reqID))
}
