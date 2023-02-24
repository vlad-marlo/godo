package production

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vlad-marlo/godo/internal/config"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/pkg/fielderr"
	"github.com/vlad-marlo/godo/internal/service"
	"github.com/vlad-marlo/godo/internal/store"
	"github.com/vlad-marlo/godo/internal/store/mocks"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"testing"
)

var (
	_user1 = &model.User{
		ID:    uuid.New(),
		Pass:  "difficult_password1",
		Email: "email@example.com",
	}
)

func TestService_RegisterUser_Positive(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := mocks.NewMockStore(ctrl)
	user := mocks.NewMockUserRepository(ctrl)
	user.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	s.EXPECT().User().Return(user).AnyTimes()
	srv := testService(t, s)

	u, err := srv.RegisterUser(context.Background(), _user1.Email, _user1.Pass)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, u.ID)
	u.ID = _user1.ID
	assert.Equal(t, _user1.Email, u.Email)
	assert.Empty(t, u.Pass)
}

func TestService_RegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := mocks.NewMockStore(ctrl)
	user := mocks.NewMockUserRepository(ctrl)
	user.EXPECT().Create(gomock.Any(), gomock.Any()).Return(store.ErrUserAlreadyExists)
	s.EXPECT().User().Return(user).AnyTimes()
	srv := testService(t, s)

	t.Run("already exists", func(t *testing.T) {
		u, err := srv.RegisterUser(context.Background(), _user1.Email, _user1.Pass)
		assert.Nil(t, u)
		assert.Error(t, err)
		require.IsType(t, &fielderr.Error{}, err)
		assert.ErrorIs(t, err, service.ErrEmailAlreadyInUse)
	})

	t.Run("too simple password", func(t *testing.T) {
		u, err := srv.RegisterUser(context.Background(), _user1.Email, "p")
		assert.Nil(t, u)
		assert.ErrorIs(t, err, service.ErrPasswordToEasy)
	})
	t.Run("to long password", func(t *testing.T) {
		u, err := srv.RegisterUser(context.Background(), _user1.Email, strings.Repeat(_user1.Pass, 10000))
		assert.Nil(t, u)
		assert.ErrorIs(t, err, service.ErrPasswordToLong)
	})
}

func TestService_RegisterUser_Unknown(t *testing.T) {
	// init necessary objects
	ctrl := gomock.NewController(t)
	s := mocks.NewMockStore(ctrl)
	user := mocks.NewMockUserRepository(ctrl)
	user.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New(""))
	s.EXPECT().User().Return(user).AnyTimes()
	srv := testService(t, s)

	u, err := srv.RegisterUser(context.Background(), _user1.Email, _user1.Pass)
	assert.Nil(t, u)
	assert.IsType(t, &fielderr.Error{}, err)
	assert.Equal(t, "internal server error", err.Error())
	fErr := err.(*fielderr.Error)
	assert.Equal(t, fielderr.CodeInternal, fErr.Code())
}

func TestService_LoginUserJWT_Positive(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := mocks.NewMockStore(ctrl)
	user := mocks.NewMockUserRepository(ctrl)

	// encrypted pass
	pass, err := bcrypt.GenerateFromPassword([]byte(config.New().Server.Salt+_user1.Pass), bcrypt.DefaultCost)
	assert.NoError(t, err)

	u1 := &model.User{ID: _user1.ID, Email: _user1.Email, Pass: string(pass)}
	user.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(u1, nil)

	s.EXPECT().User().Return(user).AnyTimes()
	srv := testService(t, s)

	resp, err := srv.CreateToken(context.Background(), _user1.Email, _user1.Pass, BearerToken)
	assert.NoError(t, err)

	assert.NotNil(t, resp)
	assert.Equal(t, resp.TokenType, "bearer")
	for _, tok := range []string{resp.AccessToken} {
		token, err := jwt.ParseWithClaims(tok, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.New().Server.SecretKey), nil
		})
		assert.NoError(t, err)
		assert.IsType(t, &jwt.RegisteredClaims{}, token.Claims)
		assert.Equal(t, u1.ID.String(), token.Claims.(*jwt.RegisteredClaims).Subject)
	}
}

func TestService_LoginUserJWT(t *testing.T) {
	tt := []struct {
		name    string
		stErr   error
		wantErr error
	}{
		{"not found", store.ErrNotFound, service.ErrBadAuthData},
		{"internal", errors.New("xd"), service.ErrInternal},
		{"unauthorized", nil, service.ErrBadAuthData},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s := mocks.NewMockStore(ctrl)
			user := mocks.NewMockUserRepository(ctrl)

			user.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(&model.User{}, tc.stErr)

			s.EXPECT().User().Return(user).AnyTimes()
			srv := testService(t, s)
			resp, err := srv.CreateToken(context.Background(), "", "", BearerToken)
			assert.Nil(t, resp)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}

}
