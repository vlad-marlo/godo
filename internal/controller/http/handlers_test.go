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
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_RegisterUser_Positive(t *testing.T) {
	var body []byte
	var err error
	req := &model.RegisterUserRequest{
		Username: TestUser1.Name,
		Password: TestUser1.Pass,
	}
	{
		body, err = json.Marshal(req)
		require.NoError(t, err)
	}
	ctrl := gomock.NewController(t)
	srv := mocks.NewMockService(ctrl)

	srv.EXPECT().RegisterUser(gomock.Any(), req.Username, req.Password, false).Return(TestUser1, nil)
	s := TestServer(t, srv)

	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.RegisterUser(w, r)

	res := w.Result()

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
				Username: TestUser1.Name,
				Password: TestUser1.Pass,
			})
			require.NoError(t, err, "prepare request data")
			body := bytes.NewReader(req)

			ctrl := gomock.NewController(t)
			srv := mocks.NewMockService(ctrl)
			srv.EXPECT().RegisterUser(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, tc.err).AnyTimes()
			s := TestServer(t, srv)

			r := httptest.NewRequest(http.MethodPost, "/", body)
			w := httptest.NewRecorder()

			s.RegisterUser(w, r)

			res := w.Result()
			fErr, ok := tc.err.(*fielderr.Error)
			if !ok {
				assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
				return
			}
			assert.Equal(t, fErr.CodeHTTP(), res.StatusCode)
			if fErr.Data == nil {
				fErr.Data = http.StatusText(fErr.CodeHTTP())
			}
			expected, err := json.Marshal(fErr.Data)
			require.NoError(t, err)
			assert.JSONEq(t, string(expected), w.Body.String())
		})
	}
}

//func TestServer_RegisterUser_NilBody(t *testing.T) {
//	s := TestServer(t, nil)
//	r := httptest.NewRequest(http.MethodPost, "/", nil)
//	w := httptest.NewRecorder()
//	s.RegisterUser(w, r)
//	assert.Equal(t, http.StatusText(http.StatusInternalServerError), w.Body.String())
//	assert.Equal(t, http.StatusInternalServerError, w.Code)
//}
