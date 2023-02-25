package httpctrl

import (
	"bytes"
	"encoding/json"
	"github.com/go-chi/chi/v5/middleware"
	mw "github.com/vlad-marlo/godo/internal/controller/http/middleware"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/pkg/fielderr"
	"go.uber.org/zap"
	"io"
	"net/http"
)

const ZapRequestIDFieldName = "request_id"

// ReqIDField return named zap field with reqID in it.
func ReqIDField(reqID string) zap.Field {
	return zap.String(ZapRequestIDFieldName, reqID)
}

// RegisterUser creates user with provided data.
//
//	@Tags		UserCreate
//	@Summary	Создание пользователя
//	@ID			user_create
//	@Accept		json
//	@Produce	json
//	@Param		request	body		model.RegisterUserRequest	true	"User data"
//	@Success	201		{object}	model.User
//	@Failure	400		{object}	model.Error	"Bad Request"
//	@Failure	409		{object}	model.Error	"Conflict"
//	@Failure	500		{object}	model.Error	"Internal Server Error"
//	@Router		/users/register [post]
func (s *Server) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterUserRequest
	reqID := middleware.GetReqID(r.Context())

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r.Body); err != nil {
		s.respond(w, http.StatusInternalServerError, nil, ReqIDField(reqID), zap.Error(err))
		return
	}

	if err := json.NewDecoder(&buf).Decode(&req); err != nil {
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

// CreateToken creates JWT bearer token with provided data.
//
//	@Tags		CreateToken
//	@Summary	Создание JWT токена для пользователя.
//	@ID			login_jwt
//	@Accept		json
//	@Produce	json
//	@Param		request	body		model.CreateTokenRequest	true	"User data"
//	@Success	201		{object}	model.CreateTokenResponse
//	@Failure	400		{object}	model.Error	"Bad Request"
//	@Failure	401		{string}	model.Error	"Unauthorized"
//	@Failure	500		{string}	model.Error	"Internal Server Error"
//	@Router		/users/token [post]
func (s *Server) CreateToken(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTokenRequest
	var buf bytes.Buffer
	reqID := middleware.GetReqID(r.Context())

	if _, err := io.Copy(&buf, r.Body); err != nil {
		s.respond(w, http.StatusInternalServerError, nil, zap.Error(err), ReqIDField(reqID))
		return
	}
	_ = r.Body.Close()

	if err := json.NewDecoder(&buf).Decode(&req); err != nil {
		s.respond(w, http.StatusBadRequest, nil, zap.Error(err), ReqIDField(reqID))
		return
	}

	u, err := s.srv.CreateToken(r.Context(), req.Email, req.Password, req.TokenType)
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
//
//	@Tags		Ping
//	@Summary	Запрос состояния сервиса
//	@ID			ping
//	@Accept		plain
//	@Produce	plain
//	@Success	200	{string}	string	"OK"
//	@Failure	500	{string}	string	"Internal Server Error"
//	@Router		/ping [get]
func (s *Server) Ping(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	if err := s.srv.Ping(r.Context()); err != nil {

		if fErr, ok := err.(*fielderr.Error); ok {
			s.respond(w, fErr.CodeHTTP(), fErr.Data(), zap.Error(err), ReqIDField(reqID))
			return
		}

		s.respond(w, http.StatusInternalServerError, nil, zap.Error(err), ReqIDField(reqID))
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CreateGroup create new group.
//
//	@tags						CreateGroup
//	@summary					Создание группы пользователей
//	@ID							group_create
//	@Accept						json
//	@produce					json
//	@Param						request	body	model.CreateGroupRequest	true	"group data"
//
// Success 201 {object} model.CreateGroupResponse
// Failure 400 {object} model.Error
// Failure 401 {object} model.Error
// Failure 409 {object} model.Error
// Failure 500 {object} model.Error
//
//	@securityDefinitions.apikey	ApiKeyAuth
func (s *Server) CreateGroup(w http.ResponseWriter, r *http.Request) {
	// usage of buffer make unnecessary deferring closing of request body. That saves about 6ns
	// source - (https://go.googlesource.com/proposal/+/refs/heads/master/design/34481-opencoded-defers.md)
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r.Body); err != nil {
		s.respond(w, http.StatusInternalServerError, nil, zap.Error(err))
		return
	}
	_ = r.Body.Close()

	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	user := mw.UserFromCtx(ctx)

	var req model.CreateGroupRequest
	if err := json.NewDecoder(&buf).Decode(&req); err != nil {
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
