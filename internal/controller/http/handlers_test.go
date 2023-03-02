package httpctrl

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vlad-marlo/godo/internal/controller/http/mocks"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/pkg/fielderr"
	"github.com/vlad-marlo/godo/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
	srv := mocks.NewMockService(ctrl)

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
			srv := mocks.NewMockService(ctrl)
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
	srv := mocks.NewMockService(ctrl)
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
	srv := mocks.NewMockService(ctrl)
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
	srv := mocks.NewMockService(ctrl)
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
	srv := mocks.NewMockService(ctrl)
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
	srv := mocks.NewMockService(ctrl)
	resp := &model.CreateTokenResponse{
		TokenType:   "authorization",
		AccessToken: "jklsadfhlsadfhjksadjlhf",
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
			srv := mocks.NewMockService(ctrl)
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
	srv := mocks.NewMockService(ctrl)
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
