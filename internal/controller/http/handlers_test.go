package httpctrl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vlad-marlo/godo/internal/config"
	mw "github.com/vlad-marlo/godo/internal/controller/http/middleware"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/pkg/fielderr"
	"github.com/vlad-marlo/godo/internal/service"
	"github.com/vlad-marlo/godo/internal/service/mocks"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestServer_RegisterUser_Positive(t *testing.T) {
	var body []byte
	var err error
	req := &model.RegisterUserRequest{
		Email:    TestUser1.Email,
		Password: TestUser1.Pass,
	}
	{
		body, err = json.Marshal(req)
		require.NoError(t, err)
	}

	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)

	srv.EXPECT().RegisterUser(gomock.Any(), req.Email, req.Password).Return(TestUser1, nil)
	s := TestServer(t, srv)

	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	defer assert.NoError(t, r.Body.Close())
	w := httptest.NewRecorder()

	s.RegisterUser(w, r)

	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	assert.Contains(t, res.Header.Get("content-type"), "application/json")
	assert.Equal(t, http.StatusCreated, res.StatusCode)
	// test response
	expected, err := json.Marshal(TestUser1)
	require.NoError(t, err)
	assert.JSONEq(t, string(expected), w.Body.String())
}

func TestServer_RegisterUser_Negative(t *testing.T) {
	tt := []struct {
		name string
		err  error
	}{
		{"internal", fielderr.New("internal", nil, fielderr.CodeInternal)},
		{"not found", fielderr.New("not found", nil, fielderr.CodeNotFound)},
		{"unauthorized", fielderr.New("unauthorized", nil, fielderr.CodeUnauthorized)},
		{"conflict", fielderr.New("conflict", nil, fielderr.CodeConflict)},
		{"conflict with data", fielderr.New("conflict", TestUser1, fielderr.CodeConflict)},
		{"unknown error", errors.New("unknown")},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			req, err := json.Marshal(&model.RegisterUserRequest{
				Email:    TestUser1.Email,
				Password: TestUser1.Pass,
			})
			require.NoError(t, err, "prepare request data")
			body := bytes.NewReader(req)

			ctrl := gomock.NewController(t)
			srv := mocks.NewMockInterface(ctrl)
			srv.EXPECT().RegisterUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, tc.err).AnyTimes()
			s := TestServer(t, srv)

			r := httptest.NewRequest(http.MethodPost, "/", body)
			defer assert.NoError(t, r.Body.Close())
			w := httptest.NewRecorder()

			s.RegisterUser(w, r)

			res := w.Result()
			defer assert.NoError(t, res.Body.Close())
			fErr, ok := tc.err.(*fielderr.Error)
			if !ok {
				assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
				return
			}
			assert.Equal(t, fErr.CodeHTTP(), res.StatusCode)
			data := fErr.Data()
			if data == nil {
				data = http.StatusText(fErr.CodeHTTP())
				t.Logf("got nil data: %v", data)
			}
			var expected []byte
			expected, err = json.Marshal(data)
			require.NoError(t, err)
			assert.JSONEq(t, string(expected), w.Body.String())
		})
	}
}

func TestServer_RegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)
	s := TestServer(t, srv)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("[xd:"))
	defer assert.NoError(t, r.Body.Close())
	w := httptest.NewRecorder()

	s.RegisterUser(w, r)
	res := w.Result()
	defer assert.NoError(t, res.Body.Close())
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	data := http.StatusText(res.StatusCode)
	t.Logf("got nil data: %v", data)
	expected, err := json.Marshal(data)
	require.NoError(t, err)
	assert.JSONEq(t, string(expected), w.Body.String())
}

func TestServer_Ping_FieldErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)
	fErr := service.ErrPasswordToLong
	srv.EXPECT().Ping(gomock.Any()).Return(fErr)
	s := TestServer(t, srv)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	s.Ping(w, r)
	defer assert.NoError(t, r.Body.Close())
	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	assert.Equal(t, fErr.CodeHTTP(), res.StatusCode)

	data, err := json.Marshal(fErr.Data())
	require.NoError(t, err)
	assert.JSONEq(t, string(data), w.Body.String())
}

func TestServer_Ping_NilErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)
	srv.EXPECT().Ping(gomock.Any()).Return(nil)
	s := TestServer(t, srv)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	s.Ping(w, r)
	defer assert.NoError(t, r.Body.Close())
	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	assert.Equal(t, "", w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestServer_Ping_UnknownErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)
	srv.EXPECT().Ping(gomock.Any()).Return(errors.New(""))
	s := TestServer(t, srv)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	s.Ping(w, r)
	defer assert.NoError(t, r.Body.Close())
	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	data, err := json.Marshal(http.StatusText(http.StatusInternalServerError))
	require.NoError(t, err)
	assert.JSONEq(t, string(data), w.Body.String())
}

func TestServer_CreateToken_MainPositive(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)
	resp := &model.CreateTokenResponse{
		TokenType:   "authorization",
		AccessToken: "some token",
	}
	srv.EXPECT().CreateToken(gomock.Any(), TestTokenRequest.Email, TestTokenRequest.Password, TestTokenRequest.TokenType).Return(resp, nil)
	s := TestServer(t, srv)

	w := httptest.NewRecorder()
	body, err := json.Marshal(TestTokenRequest)
	require.NoError(t, err)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	defer assert.NoError(t, r.Body.Close())

	s.CreateToken(w, r)

	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	want, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.JSONEq(t, string(want), w.Body.String())
}

func TestServer_CreateToken_Bad(t *testing.T) {
	body, err := json.Marshal(TestTokenRequest)
	require.NoError(t, err)

	tt := []struct {
		name   string
		srvErr error
	}{
		{"internal", service.ErrInternal},
		{"unknown", errors.New("")},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srv := mocks.NewMockInterface(ctrl)
			srv.EXPECT().CreateToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, tc.srvErr)
			s := TestServer(t, srv)

			w := httptest.NewRecorder()

			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			defer assert.NoError(t, r.Body.Close())

			s.CreateToken(w, r)

			res := w.Result()
			defer assert.NoError(t, res.Body.Close())

			if fErr, ok := tc.srvErr.(*fielderr.Error); ok {
				assert.Equal(t, fErr.CodeHTTP(), res.StatusCode)
				var data []byte
				data, err = json.Marshal(http.StatusText(fErr.CodeHTTP()))
				require.NoError(t, err)
				if fErr.Data() != nil {
					data, err = json.Marshal(fErr.Data())
				}
				assert.JSONEq(t, string(data), w.Body.String())
			} else {
				assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
				var data []byte
				data, err = json.Marshal(http.StatusText(http.StatusInternalServerError))
				require.NoError(t, err)
				assert.JSONEq(t, string(data), w.Body.String())
			}
		})
	}
}

func TestServer_CreateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)
	s := TestServer(t, srv)

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("[{d}:"))
	defer assert.NoError(t, r.Body.Close())

	s.CreateToken(w, r)

	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	data, err := json.Marshal(http.StatusText(http.StatusBadRequest))
	require.NoError(t, err)
	assert.JSONEq(t, string(data), w.Body.String())
}

func TestServer_CreateGroup_MainPositive(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)

	req := &model.CreateGroupRequest{
		Name:        "test group",
		Description: "test description",
	}
	resp := &model.CreateGroupResponse{
		ID:          uuid.Nil,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now().Unix(),
	}
	srv.EXPECT().CreateGroup(context.Background(), uuid.Nil, req.Name, req.Description).Return(resp, nil)
	s := TestServer(t, srv)
	b, err := json.Marshal(req)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	s.CreateGroup(w, r)
	defer assert.NoError(t, r.Body.Close())
	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	var got model.CreateGroupResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, *resp, got)
}

func TestServer_CreateGroup_NoData(t *testing.T) {
	s := TestServer(t, nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	//require.NoError(t, r.Body.Close())
	s.CreateGroup(w, r)
	res := w.Result()
	defer assert.NoError(t, res.Body.Close())
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

}

func TestServer_CreateGroup_UnknownErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)

	req := &model.CreateGroupRequest{
		Name:        "test group",
		Description: "test description",
	}

	srv.EXPECT().
		CreateGroup(context.Background(), uuid.Nil, req.Name, req.Description).
		Return(nil, errors.New(""))

	s := TestServer(t, srv)

	b, err := json.Marshal(req)
	require.NoError(t, err)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	defer assert.NoError(t, r.Body.Close())
	w := httptest.NewRecorder()

	s.CreateGroup(w, r)
	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestServer_CreateGroup_FieldErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)

	req := &model.CreateGroupRequest{
		Name:        "test group",
		Description: "test description",
	}
	fErr := fielderr.New("some error", req, fielderr.CodeForbidden)

	srv.EXPECT().
		CreateGroup(context.Background(), uuid.Nil, req.Name, req.Description).
		Return(nil, fErr)

	s := TestServer(t, srv)

	b, err := json.Marshal(req)
	require.NoError(t, err)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	defer assert.NoError(t, r.Body.Close())
	w := httptest.NewRecorder()

	s.CreateGroup(w, r)

	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	assert.Equal(t, fErr.CodeHTTP(), res.StatusCode)

	var jsonData []byte
	jsonData, err = json.Marshal(fErr.Data())
	require.NoError(t, err)
	assert.JSONEq(t, string(jsonData), w.Body.String())
}

func TestServer_CreateInviteLink_MainPositive(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)

	req := &model.CreateInviteRequest{
		Group:   uuid.New(),
		Limit:   2,
		Member:  3,
		Task:    4,
		Review:  5,
		Comment: 6,
	}

	role := &model.Role{
		Members:  req.Member,
		Tasks:    req.Task,
		Reviews:  req.Review,
		Comments: req.Comment,
	}

	resp := &model.CreateInviteResponse{
		Link:  "some link",
		Limit: req.Limit,
	}

	srv.EXPECT().CreateInvite(context.Background(), uuid.Nil, req.Group, gomock.Eq(role), 2).Return(resp, nil)

	s := TestServer(t, srv)

	b, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	defer assert.NoError(t, r.Body.Close())
	w := httptest.NewRecorder()

	s.CreateInviteLink(w, r)
	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	var jsonResp []byte
	jsonResp, err = json.Marshal(resp)
	require.NoError(t, err)
	assert.JSONEq(t, string(jsonResp), w.Body.String())
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestServer_CreateInviteLink_BadData(t *testing.T) {
	s := TestServer(t, nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	defer assert.NoError(t, r.Body.Close())
	s.CreateInviteLink(w, r)
	res := w.Result()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestServer_CreateInviteLink_DifferentErrors(t *testing.T) {
	type testCase struct {
		name string
		err  error
	}
	tt := []testCase{
		{"unknown", errors.New("")},
		{"bad auth data", service.ErrBadAuthData},
		{"internal", service.ErrBadInvite},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srv := mocks.NewMockInterface(ctrl)

			req := &model.CreateInviteRequest{
				Group:   uuid.New(),
				Limit:   2,
				Member:  3,
				Task:    4,
				Review:  5,
				Comment: 6,
			}

			role := &model.Role{
				Members:  req.Member,
				Tasks:    req.Task,
				Reviews:  req.Review,
				Comments: req.Comment,
			}

			srv.EXPECT().CreateInvite(context.Background(), uuid.Nil, req.Group, gomock.Eq(role), 2).Return(nil, tc.err)

			s := TestServer(t, srv)

			b, err := json.Marshal(req)
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
			defer assert.NoError(t, r.Body.Close())
			w := httptest.NewRecorder()

			s.CreateInviteLink(w, r)
			res := w.Result()
			defer assert.NoError(t, res.Body.Close())

			fErr, ok := tc.err.(*fielderr.Error)
			if !ok {
				require.Equal(t, http.StatusInternalServerError, w.Code)
				return
			}
			var data []byte
			data, err = json.Marshal(fErr.Data())
			require.NoError(t, err)
			assert.JSONEq(t, string(data), w.Body.String())
		})
	}
}

func TestServer_CreateInviteViaGroup_MainPositive(t *testing.T) {
	id := uuid.New()
	req := &model.CreateInviteViaGroupRequest{
		Limit:   2,
		Member:  2,
		Task:    0,
		Review:  2,
		Comment: 2,
	}

	b, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	r = reqWithGroup(t, r, id.String())
	r = mw.RequestWithUser(r, id)
	defer assert.NoError(t, r.Body.Close())
	w := httptest.NewRecorder()

	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)
	resp := &model.CreateInviteResponse{
		Link:  "some link",
		Limit: 2,
	}
	srv.
		EXPECT().
		CreateInvite(
			gomock.Any(),
			id,
			id,
			gomock.Eq(&model.Role{
				ID:       0,
				Members:  req.Member,
				Tasks:    req.Task,
				Reviews:  req.Review,
				Comments: req.Comment,
			}),
			req.Limit,
		).
		Return(resp, nil)

	s := TestServer(t, srv)

	s.CreateInviteViaGroup(w, r)

	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	var expected []byte
	expected, err = json.Marshal(resp)
	require.NoError(t, err)

	assert.JSONEq(t, string(expected), w.Body.String())
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestServer_CreateInviteViaGroup_BadGroup(t *testing.T) {
	id := uuid.New()
	req := &model.CreateInviteViaGroupRequest{
		Limit:   2,
		Member:  2,
		Task:    0,
		Review:  2,
		Comment: 2,
	}

	b, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	r = reqWithGroup(t, r, "bad_id")
	r = mw.RequestWithUser(r, id)
	defer assert.NoError(t, r.Body.Close())
	w := httptest.NewRecorder()

	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)

	s := TestServer(t, srv)

	s.CreateInviteViaGroup(w, r)

	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	var expected []byte
	expected, err = json.Marshal(map[string]string{"path": "bad group id"})
	require.NoError(t, err)

	assert.JSONEq(t, string(expected), w.Body.String())
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServer_CreateInviteViaGroup_Errors(t *testing.T) {
	tt := []struct {
		name string
		err  error
	}{
		{"unknown error", errors.New("")},
		{"field error: conflict", fielderr.New("some msg", map[string]string{"some": "data"}, fielderr.CodeConflict)},
		{"field error: internal", fielderr.New("some msg", "some text", fielderr.CodeInternal)},
		{"field error: forbidden", fielderr.New("some msg", config.New(), fielderr.CodeForbidden)},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			id := uuid.New()
			req := &model.CreateInviteViaGroupRequest{
				Limit:   2,
				Member:  2,
				Task:    0,
				Review:  2,
				Comment: 2,
			}

			b, err := json.Marshal(req)
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
			r = reqWithGroup(t, r, id.String())
			r = mw.RequestWithUser(r, id)
			defer assert.NoError(t, r.Body.Close())
			w := httptest.NewRecorder()

			ctrl := gomock.NewController(t)
			srv := mocks.NewMockInterface(ctrl)
			srv.
				EXPECT().
				CreateInvite(
					gomock.Any(),
					id,
					id,
					gomock.Eq(&model.Role{
						ID:       0,
						Members:  req.Member,
						Tasks:    req.Task,
						Reviews:  req.Review,
						Comments: req.Comment,
					}),
					req.Limit,
				).
				Return(nil, tc.err)

			s := TestServer(t, srv)

			s.CreateInviteViaGroup(w, r)

			res := w.Result()
			defer assert.NoError(t, res.Body.Close())

			fErr, ok := tc.err.(*fielderr.Error)
			if !ok {
				assert.Equal(t, http.StatusInternalServerError, w.Code)
				return
			}
			var expected []byte
			expected, err = json.Marshal(fErr.Data())
			require.NoError(t, err)

			assert.JSONEq(t, string(expected), w.Body.String())
			assert.Equal(t, fErr.CodeHTTP(), w.Code)
		})
	}
}

func TestServer_UseInvite_Positive(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)

	invite := uuid.New()
	group := uuid.New()
	srv.EXPECT().UseInvite(gomock.Any(), uuid.Nil, group, invite).Return(nil)

	s := TestServer(t, srv)

	r := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/?%s=%s", InviteInQueryKey, invite.String()), nil)
	r = reqWithGroup(t, r, group.String())
	defer assert.NoError(t, r.Body.Close())
	w := httptest.NewRecorder()

	s.UseInvite(w, r)
	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestServer_UseInvite_BadInvite(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)

	invite := "xd"
	group := uuid.New()

	s := TestServer(t, srv)

	r := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/?%s=%s", InviteInQueryKey, invite), nil)
	r = reqWithGroup(t, r, group.String())
	defer assert.NoError(t, r.Body.Close())
	w := httptest.NewRecorder()

	s.UseInvite(w, r)
	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	data, err := json.Marshal(map[string]string{"query": "invite must be valid uuid"})
	require.NoError(t, err)
	assert.JSONEq(t, string(data), w.Body.String())
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestServer_UseInvite_BadGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)

	invite := uuid.New()
	group := "bad_group_id"

	s := TestServer(t, srv)

	r := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/?%s=%s", InviteInQueryKey, invite.String()), nil)
	r = reqWithGroup(t, r, group)
	defer assert.NoError(t, r.Body.Close())
	w := httptest.NewRecorder()

	s.UseInvite(w, r)
	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	data, err := json.Marshal(map[string]string{"url": "invite must be valid group id in it"})
	require.NoError(t, err)
	assert.JSONEq(t, string(data), w.Body.String())
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestServer_UseInvite_NegativeBadErrors(t *testing.T) {
	tt := []struct {
		name string
		err  error
	}{
		{"unknown error", errors.New("")},
		{"field error: conflict", fielderr.New("some msg", map[string]string{"some": "data"}, fielderr.CodeConflict)},
		{"field error: internal", fielderr.New("some msg", "some text", fielderr.CodeInternal)},
		{"field error: forbidden", fielderr.New("some msg", config.New(), fielderr.CodeForbidden)},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srv := mocks.NewMockInterface(ctrl)

			invite := uuid.New()
			group := uuid.New()
			srv.EXPECT().UseInvite(gomock.Any(), uuid.Nil, group, invite).Return(tc.err)

			s := TestServer(t, srv)

			r := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/?%s=%s", InviteInQueryKey, invite.String()), nil)
			r = reqWithGroup(t, r, group.String())
			defer assert.NoError(t, r.Body.Close())
			w := httptest.NewRecorder()

			s.UseInvite(w, r)
			res := w.Result()
			defer assert.NoError(t, res.Body.Close())

			fErr, ok := tc.err.(*fielderr.Error)
			if !ok {
				assert.Equal(t, http.StatusInternalServerError, w.Code)
				return
			}
			body, err := json.Marshal(fErr.Data())
			require.NoError(t, err)

			assert.Equal(t, fErr.CodeHTTP(), w.Code)
			assert.JSONEq(t, string(body), w.Body.String())
		})
	}
}

func TestServer_UserMe_Positive(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)

	u := uuid.New()
	resp := &model.GetMeResponse{
		ID:    u,
		Email: "example@email.org",
		Groups: []model.GroupInUser{
			{
				uuid.New(),
				"group #1",
				"description",
				[]*model.Task{
					{uuid.New(), "task", "description", time.Now(), u, time.Now().Unix(), "NEW"},
					{uuid.New(), "other task", "other description", time.Now(), u, time.Now().Unix(), "NEW"},
				},
			},
			{uuid.New(), "group #2", "other desc", nil},
		},
	}

	srv.EXPECT().GetMe(gomock.Any(), u).Return(resp, nil)

	r := mw.RequestWithUser(httptest.NewRequest("", "/", nil), u)
	defer assert.NoError(t, r.Body.Close())
	w := httptest.NewRecorder()
	s := TestServer(t, srv)

	s.UserMe(w, r)

	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	body, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.JSONEq(t, string(body), w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestServer_UserMe_Negative(t *testing.T) {
	tt := []struct {
		name string
		err  error
	}{
		{"unknown error", errors.New("")},
		{"field error: conflict", fielderr.New("some msg", map[string]string{"some": "data"}, fielderr.CodeConflict)},
		{"field error: internal", fielderr.New("some msg", "some text", fielderr.CodeInternal)},
		{"field error: forbidden", fielderr.New("some msg", config.New(), fielderr.CodeForbidden)},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srv := mocks.NewMockInterface(ctrl)
			srv.EXPECT().GetMe(gomock.Any(), uuid.Nil).Return(nil, tc.err)
			s := TestServer(t, srv)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("", "/", nil)
			defer assert.NoError(t, r.Body.Close())

			s.UserMe(w, r)

			res := w.Result()
			defer assert.NoError(t, res.Body.Close())

			fErr, ok := tc.err.(*fielderr.Error)
			if !ok {
				assert.Equal(t, http.StatusInternalServerError, w.Code)
				return
			}
			body, err := json.Marshal(fErr.Data())
			require.NoError(t, err)

			assert.JSONEq(t, string(body), w.Body.String())
			assert.Equal(t, fErr.CodeHTTP(), w.Code)
		})
	}
}

func TestServer_GetTask_MainPositive(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)
	resp := &model.Task{
		ID:          uuid.New(),
		Name:        "task name",
		Description: "task desc",
		CreatedAt:   time.Now(),
		CreatedBy:   uuid.Nil,
		Created:     0,
		Status:      "NEW",
	}
	srv.EXPECT().GetTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(resp, nil)

	s := TestServer(t, srv)

	w := httptest.NewRecorder()
	r := reqWithTask(t, httptest.NewRequest("", "/", nil), uuid.NewString())
	defer assert.NoError(t, r.Body.Close())

	s.GetTask(w, r)

	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	body, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.JSONEq(t, string(body), w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestServer_GetTask_BadTask(t *testing.T) {
	s := TestServer(t, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("", "/", nil)
	defer assert.NoError(t, r.Body.Close())

	s.GetTask(w, r)

	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServer_GetTask_Errors(t *testing.T) {
	tt := []struct {
		name string
		err  error
	}{
		{"unknown error", errors.New("")},
		{"field error: conflict", fielderr.New("some msg", map[string]string{"some": "data"}, fielderr.CodeConflict)},
		{"field error: internal", fielderr.New("some msg", "some text", fielderr.CodeInternal)},
		{"field error: forbidden", fielderr.New("some msg", config.New(), fielderr.CodeForbidden)},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srv := mocks.NewMockInterface(ctrl)

			srv.
				EXPECT().
				GetTask(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, tc.err)

			s := TestServer(t, srv)

			w := httptest.NewRecorder()
			r := reqWithTask(t, httptest.NewRequest("", "/", nil), uuid.NewString())
			defer assert.NoError(t, r.Body.Close())

			s.GetTask(w, r)

			fErr, ok := tc.err.(*fielderr.Error)
			if !ok {
				assert.Equal(t, http.StatusInternalServerError, w.Code)
				return
			}
			data, err := json.Marshal(fErr.Data())
			require.NoError(t, err)
			assert.JSONEq(t, string(data), w.Body.String())
			assert.Equal(t, fErr.CodeHTTP(), w.Code)
		})
	}
}

func TestServer_AllTasks_Positive(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)
	resp := &model.GetTasksResponse{
		Count: 5,
		Tasks: []*model.Task{
			{uuid.New(), uuid.NewString(), uuid.NewString(), time.Now(), uuid.New(), 0, uuid.NewString()},
			{uuid.New(), uuid.NewString(), uuid.NewString(), time.Now(), uuid.New(), 0, uuid.NewString()},
			{uuid.New(), uuid.NewString(), uuid.NewString(), time.Now(), uuid.New(), 0, uuid.NewString()},
			{uuid.New(), uuid.NewString(), uuid.NewString(), time.Now(), uuid.New(), 0, uuid.NewString()},
			{uuid.New(), uuid.NewString(), uuid.NewString(), time.Now(), uuid.New(), 0, uuid.NewString()},
		},
	}
	srv.EXPECT().GetUserTasks(gomock.Any(), uuid.Nil).Return(resp, nil)

	s := TestServer(t, srv)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("", "/", nil)
	defer assert.NoError(t, r.Body.Close())

	s.AllTasks(w, r)

	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	body, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.JSONEq(t, string(body), w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestServer_AllTasks_Errors(t *testing.T) {
	tt := []struct {
		name string
		err  error
	}{
		{"unknown error", errors.New("")},
		{"field error: conflict", fielderr.New("some msg", map[string]string{"some": "data"}, fielderr.CodeConflict)},
		{"field error: internal", fielderr.New("some msg", "some text", fielderr.CodeInternal)},
		{"field error: forbidden", fielderr.New("some msg", config.New(), fielderr.CodeForbidden)},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srv := mocks.NewMockInterface(ctrl)

			srv.
				EXPECT().
				GetUserTasks(gomock.Any(), gomock.Any()).
				Return(nil, tc.err)

			s := TestServer(t, srv)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("", "/", nil)
			defer assert.NoError(t, r.Body.Close())

			s.AllTasks(w, r)

			fErr, ok := tc.err.(*fielderr.Error)
			if !ok {
				assert.Equal(t, http.StatusInternalServerError, w.Code)
				return
			}
			data, err := json.Marshal(fErr.Data())
			require.NoError(t, err)
			assert.JSONEq(t, string(data), w.Body.String())
			assert.Equal(t, fErr.CodeHTTP(), w.Code)
		})
	}
}

func TestServer_CreateTask_Positive(t *testing.T) {
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockInterface(ctrl)

	req := model.TaskCreateRequest{
		Name:        uuid.NewString(),
		Description: uuid.NewString(),
		Users:       []uuid.UUID{uuid.New(), uuid.New(), uuid.New()},
		Group:       uuid.New(),
	}

	task := &model.Task{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
		CreatedBy:   uuid.Nil,
		Created:     0,
		Status:      "NEW",
	}

	srv.EXPECT().CreateTask(gomock.Any(), uuid.Nil, req).Return(task, nil)

	s := TestServer(t, srv)

	body, err := json.Marshal(req)
	require.NoError(t, err)
	r := httptest.NewRequest("", "/", bytes.NewReader(body))
	defer assert.NoError(t, r.Body.Close())
	w := httptest.NewRecorder()

	s.CreateTask(w, r)
	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	var exp []byte
	exp, err = json.Marshal(task)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.JSONEq(t, string(exp), w.Body.String())
}

func TestServer_CreateTask_BadRequest(t *testing.T) {
	s := TestServer(t, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("", "/", nil)
	defer assert.NoError(t, r.Body.Close())

	s.CreateTask(w, r)

	res := w.Result()
	defer assert.NoError(t, res.Body.Close())

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServer_CreateTask_BadErr(t *testing.T) {
	tt := []struct {
		name string
		err  error
	}{
		{"unknown error", errors.New("")},
		{"field error: conflict", fielderr.New("some msg", map[string]string{"some": "data"}, fielderr.CodeConflict)},
		{"field error: internal", fielderr.New("some msg", "some text", fielderr.CodeInternal)},
		{"field error: forbidden", fielderr.New("some msg", config.New(), fielderr.CodeForbidden)},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srv := mocks.NewMockInterface(ctrl)

			req := model.TaskCreateRequest{
				Name:        uuid.NewString(),
				Description: uuid.NewString(),
				Users:       []uuid.UUID{uuid.New(), uuid.New(), uuid.New()},
				Group:       uuid.New(),
			}

			srv.EXPECT().CreateTask(gomock.Any(), uuid.Nil, req).Return(nil, tc.err)

			s := TestServer(t, srv)

			body, err := json.Marshal(req)
			require.NoError(t, err)

			r := httptest.NewRequest("", "/", bytes.NewReader(body))
			defer assert.NoError(t, r.Body.Close())
			w := httptest.NewRecorder()

			s.CreateTask(w, r)

			fErr, ok := tc.err.(*fielderr.Error)
			if !ok {
				assert.Equal(t, http.StatusInternalServerError, w.Code)
				return
			}
			data := fErr.Data()
			if data == nil {
				data = http.StatusText(fErr.CodeHTTP())
			}
			body, err = json.Marshal(data)
			require.NoError(t, err)
			assert.Equal(t, fErr.CodeHTTP(), w.Code)
			assert.JSONEq(t, string(body), w.Body.String())
		})
	}
}
