package httpctrl

import (
	"bytes"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
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
//	@Tags		User
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
//	@Tags		Tokens
//	@Summary	Создание JWT токена для пользователя.
//	@ID			login_jwt
//	@Accept		json
//	@Produce	json
//	@Param		request	body		model.CreateTokenRequest	true	"User data"
//	@Success	201		{object}	model.CreateTokenResponse
//	@Failure	400		{object}	model.Error	"Bad Request"
//	@Failure	401		{object}	model.Error	"Unauthorized"
//	@Failure	500		{object}	model.Error	"Internal Server Error"
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
	s.respond(w, http.StatusCreated, u)
}

// Ping godoc.
//
//	@Tags		Server
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
//	@Tags		Groups
//	@Summary	Создание группы пользователей
//	@ID			group_create
//	@Accept		json
//	@Produce	json
//	@Param		request	body		model.CreateGroupRequest	true	"group data"
//	@Success	201		{object}	model.CreateGroupResponse	"Created"
//	@Failure	400		{object}	model.Error					"Bad Request"
//	@Failure	401		{object}	model.Error					"Not Authorized"
//	@Failure	409		{object}	model.Error					"Conflict"
//	@Failure	500		{object}	model.Error					"Internal Server Error"
//	@Router		/groups/ [post]
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

// CreateInviteLink create new invite link.
//
//	@Tags		Invites,Groups
//	@Summary	создание приглашения в группу.
//	@ID			invite_user
//	@Accept		json
//	@Produce	json
//	@Param		request	body		model.CreateInviteRequest	true	"invite data"
//
//	@Success	201		{object}	model.CreateInviteResponse
//	@Failure	400		{object}	model.Error
//	@Failure	401		{object}	model.Error
//	@Failure	403		{object}	model.Error
//	@Failure	409		{object}	model.Error
//	@Failure	500		{object}	model.Error
//
//	@Router		/invites [post]
func (s *Server) CreateInviteLink(w http.ResponseWriter, r *http.Request) {
	u := mw.UserFromCtx(r.Context())
	reqID := middleware.GetReqID(r.Context())
	var req model.CreateInviteRequest
	var buf bytes.Buffer

	if _, err := io.Copy(&buf, r.Body); err != nil {
		s.respond(w, http.StatusInternalServerError, nil, zap.Error(err), ReqIDField(reqID))
		return
	}
	_ = r.Body.Close()

	if err := json.NewDecoder(&buf).Decode(&req); err != nil {
		s.respond(w, http.StatusInternalServerError, nil, zap.Error(err), ReqIDField(reqID))
	}

	role := &model.Role{
		Members:  req.Member,
		Tasks:    req.Task,
		Reviews:  req.Review,
		Comments: req.Comment,
	}

	resp, err := s.srv.CreateInvite(r.Context(), u, req.Group, role, req.Limit)
	if err != nil {

		if fErr, ok := err.(*fielderr.Error); ok {
			s.respond(w, fErr.CodeHTTP(), fErr.Data(), append(fErr.Fields(), ReqIDField(reqID))...)
			return
		}

		s.respond(w, http.StatusInternalServerError, nil, zap.Error(err), ReqIDField(reqID))
		return
	}

	s.respond(w, http.StatusCreated, resp)
}

// CreateInviteViaGroup create new invite link.
//
//	@Tags		Invites,Groups
//	@Summary	создание приглашения в группу.
//	@ID			invite_user_groups
//	@Accept		json
//	@Produce	json
//	@Param		request		body		model.CreateInviteViaGroupRequest	true	"invite data"
//	@Param		group_id	path		string								true	"group id"
//
//	@Success	201			{object}	model.CreateInviteResponse
//	@Failure	400			{object}	model.Error
//	@Failure	401			{object}	model.Error
//	@Failure	403			{object}	model.Error
//	@Failure	409			{object}	model.Error
//	@Failure	500			{object}	model.Error
//
//	@Router		/groups/{group_id}/invite [post]
func (s *Server) CreateInviteViaGroup(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	u := mw.UserFromCtx(r.Context())

	var req model.CreateInviteViaGroupRequest
	var buf bytes.Buffer

	if _, err := io.Copy(&buf, r.Body); err != nil {
		s.respond(w, http.StatusInternalServerError, nil, zap.Error(err), ReqIDField(reqID))
		return
	}
	_ = r.Body.Close()

	group, err := uuid.Parse(chi.URLParam(r, "group_id"))
	if err != nil {
		s.respond(w, http.StatusBadRequest, "bad group id", zap.Error(err), ReqIDField(reqID))
		return
	}

	if err = json.NewDecoder(&buf).Decode(&req); err != nil {
		s.respond(w, http.StatusInternalServerError, nil, zap.Error(err), ReqIDField(reqID))
	}

	role := &model.Role{
		Members:  req.Member,
		Tasks:    req.Task,
		Reviews:  req.Review,
		Comments: req.Comment,
	}

	var resp *model.CreateInviteResponse
	resp, err = s.srv.CreateInvite(r.Context(), u, group, role, req.Limit)
	if err != nil {
		if fErr, ok := err.(*fielderr.Error); ok {
			s.respond(w, fErr.CodeHTTP(), fErr.Data(), append(fErr.Fields(), ReqIDField(reqID))...)
			return
		}

		s.respond(w, http.StatusInternalServerError, nil, zap.Error(err), ReqIDField(reqID))
		return
	}

	s.respond(w, http.StatusCreated, resp, ReqIDField(reqID))

}

// UseInvite add user to group if invite is good.
//
//	@Tags		Groups
//	@Summary	Использование приглашения в группу.
//	@ID			apply_user_to_group
//	@Accept		plain
//	@Produce	json
//	@Param		group_id	path		string	true	"group id"
//	@Param		invite		query		string	true	"invite id"
//
//	@Success	201			{string}	model.CreateInviteResponse
//	@Failure	400			{object}	model.Error
//	@Failure	401			{object}	model.Error
//	@Failure	403			{object}	model.Error
//	@Failure	409			{object}	model.Error
//	@Failure	500			{object}	model.Error
//
//	@Router		/groups/{group_id}/apply [post]
func (s *Server) UseInvite(w http.ResponseWriter, r *http.Request) {
	reqID := ReqIDField(middleware.GetReqID(r.Context()))

	user := mw.UserFromCtx(r.Context())
	group, err := uuid.Parse(chi.URLParam(r, "group_id"))
	if err != nil {
		s.internal(w, reqID, zap.Error(err))
		return
	}
	var invite uuid.UUID
	invite, err = uuid.Parse(r.URL.Query().Get("invite"))
	if err != nil {
		s.internal(w, reqID, zap.Error(err))
		return
	}

	if err = s.srv.UseInvite(r.Context(), user, group, invite); err != nil {
		if fErr, ok := err.(*fielderr.Error); ok {
			s.respond(w, fErr.CodeHTTP(), fErr.Data(), append(fErr.Fields(), reqID)...)
			return
		}
		s.internal(w, reqID, zap.Error(err))
		return
	}
	s.respond(w, http.StatusOK, nil, reqID)
}

// UserMe return user profile.
//
//	@Tags		Users
//	@Summary	Get summary info about user.
//	@ID			users_me
//	@Accept		plain
//	@Produce	json
//
//	@Success	200	{object}	model.GetMeResponse
//	@Failure	401	{object}	model.Error
//	@Failure	404	{object}	model.Error
//	@Failure	500	{object}	model.Error
//
//	@Router		/users/me [get]
func (s *Server) UserMe(w http.ResponseWriter, r *http.Request) {
	reqID := ReqIDField(middleware.GetReqID(r.Context()))

	u := mw.UserFromCtx(r.Context())
	resp, err := s.srv.GetMe(r.Context(), u)
	if err != nil {
		if fErr, ok := err.(*fielderr.Error); ok {
			s.respond(w, fErr.CodeHTTP(), fErr.Data(), append(fErr.Fields(), reqID)...)
			return
		}
		s.respond(w, http.StatusInternalServerError, nil, reqID, zap.Error(err))
		return
	}
	s.respond(w, http.StatusOK, resp, reqID)
}

// AllTasks godoc.
//
//	@Tags		Tasks
//	@Summary	Get user tasks.
//	@ID			get_tasks
//	@Accept		plain
//	@Produce	json
//
//	@Success	200	{object}	model.GetTasksResponse
//	@Failure	401	{object}	model.Error
//	@Failure	404	{object}	model.Error
//	@Failure	500	{object}	model.Error
//
//	@Router		/tasks [get]
func (s *Server) AllTasks(w http.ResponseWriter, r *http.Request) {
	_ = r.Body.Close()
	reqID := ReqIDField(middleware.GetReqID(r.Context()))
	u := mw.UserFromCtx(r.Context())

	resp, err := s.srv.GetUserTasks(r.Context(), u)
	if err != nil {
		if fErr, ok := err.(*fielderr.Error); ok {
			s.respond(w, fErr.CodeHTTP(), fErr.Data(), append(fErr.Fields(), reqID)...)
			return
		}
		s.respond(w, http.StatusInternalServerError, nil, zap.Error(err), reqID)
		return
	}

	s.respond(w, http.StatusOK, resp, reqID)
}

// GetTask return summary info about task.
//
//	@Tags		Tasks
//	@Summary	Get task by ID.
//	@ID			get_task
//	@Accept		plain
//	@Produce	json
//	@Param		task_id	path		string	true	"task id"
//
//	@Success	200		{object}	model.Task
//	@Failure	400		{object}	model.Error
//	@Failure	401		{object}	model.Error
//	@Failure	403		{object}	model.Error
//	@Failure	409		{object}	model.Error
//	@Failure	500		{object}	model.Error
//
//	@Router		/tasks/{task_id} [get]
func (s *Server) GetTask(w http.ResponseWriter, r *http.Request) {
	reqID := ReqIDField(middleware.GetReqID(r.Context()))
	u := mw.UserFromCtx(r.Context())
	g, err := uuid.Parse(chi.URLParam(r, "task_id"))
	if err != nil {
		s.respond(w, http.StatusNotFound, nil, zap.Error(err))
		return
	}

	var resp *model.Task
	resp, err = s.srv.GetTask(r.Context(), u, g)
	if err != nil {


		if fErr, ok := err.(*fielderr.Error); ok {
			s.respond(w, fErr.CodeHTTP(), fErr.Data(), append(fErr.Fields(), reqID)...)
			return
		}

		s.respond(w, http.StatusInternalServerError, nil, reqID, zap.Error(err))
		return
	}

	s.respond(w, http.StatusOK, resp, reqID)
}

// CreateTask create task.
//
//	@Tags		Tasks
//	@Summary	Create task.
//	@ID			create_task
//	@Accept		json
//	@Produce	json
//	@Param		request	body		model.TaskCreateRequest	true	"task"
//
//	@Success	200		{object}	model.Task
//	@Failure	400		{object}	model.Error
//	@Failure	401		{object}	model.Error
//	@Failure	403		{object}	model.Error
//	@Failure	409		{object}	model.Error
//	@Failure	500		{object}	model.Error
//
//	@Router		/tasks/ [post]
func (s *Server) CreateTask(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	reqID := ReqIDField(middleware.GetReqID(r.Context()))
	if _, err := io.Copy(&buf, r.Body); err != nil {
		s.respond(w, http.StatusInternalServerError, nil, zap.Error(err), reqID)
		return
	}
	_ = r.Body.Close()

	u := mw.UserFromCtx(r.Context())

	var req model.TaskCreateRequest
	if err := json.NewDecoder(&buf).Decode(&req); err != nil {
		s.respond(w, http.StatusBadRequest, nil, zap.Error(err), reqID)
	}

	resp, err := s.srv.CreateTask(r.Context(), u, req)
	if err != nil {
		if fErr, ok := err.(*fielderr.Error); ok {
			s.respond(w, fErr.CodeHTTP(), fErr.Data(), append(fErr.Fields(), reqID)...)
			return
		}
		s.respond(w, http.StatusInternalServerError, nil, zap.Error(err), reqID)
		return
	}
	s.respond(w, http.StatusCreated, resp, reqID)
}
