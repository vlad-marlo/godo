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
	TestRole1 = &model.Role{
		Members:  model.PermChangeAll,
		Tasks:    model.PermChangeAll,
		Reviews:  model.PermChangeAll,
		Comments: model.PermChangeAll,
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

func TestService_CreateInvite_Negative_BadData(t *testing.T) {
	s := testService(t, nil)
	resp, err := s.CreateInvite(context.Background(), uuid.Nil, uuid.Nil, nil, -1)
	assert.Nil(t, resp)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, service.ErrBadInviteLimit)
	}
	resp, err = s.CreateInvite(context.Background(), uuid.New(), uuid.New(), nil, 123)
	assert.Nil(t, resp)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, service.ErrBadData)
	}
}

func TestService_CreateInvite_Positive(t *testing.T) {
	ctrl := gomock.NewController(t)
	str := mocks.NewMockStore(ctrl)
	groupRepo := mocks.NewMockGroupRepository(ctrl)
	roleRepo := mocks.NewMockRoleRepository(ctrl)
	inviteRepo := mocks.NewMockInviteRepository(ctrl)

	groupRepo.EXPECT().GetRoleOfMember(gomock.Any(), TestUser1.ID, TestGroup1.ID).Return(TestRole1, nil)
	roleRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil)
	inviteRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	str.EXPECT().Group().Return(groupRepo)
	str.EXPECT().Role().Return(roleRepo)
	str.EXPECT().Invite().Return(inviteRepo)

	s := testService(t, str)
	resp, err := s.CreateInvite(context.Background(), TestUser1.ID, TestGroup1.ID, TestRole1, 4)
	require.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, 4, resp.Limit)
		assert.NotNil(t, resp.Link)
	}
}

func TestService_CreateInvite_ErrHasNoRights(t *testing.T) {
	ctrl := gomock.NewController(t)

	str := mocks.NewMockStore(ctrl)
	grp := mocks.NewMockGroupRepository(ctrl)

	grp.EXPECT().GetRoleOfMember(gomock.Any(), TestUser1.ID, TestGroup1.ID).Return(ReadOnlyRole, nil)

	str.EXPECT().Group().Return(grp)

	s := testService(t, str)
	resp, err := s.CreateInvite(context.Background(), TestUser1.ID, TestGroup1.ID, ReadOnlyRole, 10)
	assert.Nil(t, resp)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, service.ErrForbidden)
	}
}

func TestService_CreateInvite_ErrWhileGettingRole(t *testing.T) {
	tt := []struct {
		name string
		err  error
		want error
	}{
		{"unknown", errors.New(""), service.ErrInternal},
		{"not found", store.ErrNotFound, service.ErrForbidden},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			str := mocks.NewMockStore(ctrl)
			grp := mocks.NewMockGroupRepository(ctrl)

			grp.EXPECT().GetRoleOfMember(gomock.Any(), TestUser1.ID, TestGroup1.ID).Return(nil, tc.err)

			str.EXPECT().Group().Return(grp)

			s := testService(t, str)
			resp, err := s.CreateInvite(context.Background(), TestUser1.ID, TestGroup1.ID, ReadOnlyRole, 10)
			assert.Nil(t, resp)
			if assert.Error(t, err) {
				assert.ErrorIs(t, err, tc.want)
			}
		})
	}
}

func TestService_CreateInvite_ErrWhileGettingRoleForInvite(t *testing.T) {
	ctrl := gomock.NewController(t)

	str := mocks.NewMockStore(ctrl)
	groupRepository := mocks.NewMockGroupRepository(ctrl)
	roleRepository := mocks.NewMockRoleRepository(ctrl)

	roleRepository.EXPECT().Get(gomock.Any(), ReadOnlyRole).Return(errors.New(""))
	groupRepository.EXPECT().GetRoleOfMember(gomock.Any(), TestUser1.ID, TestGroup1.ID).Return(SudoRole, nil)

	str.EXPECT().Group().Return(groupRepository)
	str.EXPECT().Role().Return(roleRepository)

	s := testService(t, str)
	resp, err := s.CreateInvite(context.Background(), TestUser1.ID, TestGroup1.ID, ReadOnlyRole, 10)
	assert.Nil(t, resp)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, service.ErrInternal)
	}
}

func TestService_CreateInvite_ErrWhileStoringInvite(t *testing.T) {
	tt := []struct {
		name string
		err  error
		want error
	}{
		{"unknown", errors.New(""), service.ErrInternal},
		{"fk violation", store.ErrFKViolation, service.ErrBadData},
		{"unique violation", store.ErrUniqueViolation, service.ErrConflict},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			str := mocks.NewMockStore(ctrl)
			groupRepository := mocks.NewMockGroupRepository(ctrl)
			roleRepository := mocks.NewMockRoleRepository(ctrl)
			inviteRepository := mocks.NewMockInviteRepository(ctrl)

			roleRepository.EXPECT().Get(gomock.Any(), ReadOnlyRole).Return(nil)
			groupRepository.EXPECT().GetRoleOfMember(gomock.Any(), TestUser1.ID, TestGroup1.ID).Return(SudoRole, nil)
			inviteRepository.EXPECT().Create(gomock.Any(), gomock.Any(), ReadOnlyRole.ID, TestGroup1.ID, 10).Return(tc.err)

			str.EXPECT().Group().Return(groupRepository)
			str.EXPECT().Role().Return(roleRepository)
			str.EXPECT().Invite().Return(inviteRepository)

			s := testService(t, str)
			resp, err := s.CreateInvite(context.Background(), TestUser1.ID, TestGroup1.ID, ReadOnlyRole, 10)
			assert.Nil(t, resp)
			if assert.Error(t, err) {
				assert.ErrorIs(t, err, tc.want)
			}
		})
	}
}

func TestService_GetMe_ErrWhileGetUser(t *testing.T) {
	tt := []struct {
		name string
		err  error
		want error
	}{
		{"unknown", errors.New(""), service.ErrInternal},
		{"not found", store.ErrNotFound, service.ErrUserNotFound},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			userRepo := mocks.NewMockUserRepository(ctrl)
			str := mocks.NewMockStore(ctrl)

			userRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, tc.err)

			str.EXPECT().User().Return(userRepo)

			s := testService(t, str)
			resp, err := s.GetMe(context.Background(), uuid.New())
			assert.Nil(t, resp)
			if assert.Error(t, err) {
				assert.ErrorIs(t, err, tc.want)
			}
		})
	}
}

func TestService_GetMe_ErrWhileGetByUser_Unknown(t *testing.T) {
	ctrl := gomock.NewController(t)
	str := mocks.NewMockStore(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	groupRepo := mocks.NewMockGroupRepository(ctrl)

	userRepo.EXPECT().Get(gomock.Any(), TestUser1.ID).Return(TestUser1, nil)
	groupRepo.EXPECT().GetByUser(gomock.Any(), TestUser1.ID).Return(nil, errors.New(""))

	str.EXPECT().User().Return(userRepo)
	str.EXPECT().Group().Return(groupRepo)

	s := testService(t, str)
	resp, err := s.GetMe(context.Background(), TestUser1.ID)
	assert.Nil(t, resp)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, service.ErrInternal)
	}
}

func TestService_GetMe_ErrWhileGetByUser_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	str := mocks.NewMockStore(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	groupRepo := mocks.NewMockGroupRepository(ctrl)

	userRepo.EXPECT().Get(gomock.Any(), TestUser1.ID).Return(TestUser1, nil)
	groupRepo.EXPECT().GetByUser(gomock.Any(), TestUser1.ID).Return(nil, store.ErrNotFound)

	str.EXPECT().User().Return(userRepo)
	str.EXPECT().Group().Return(groupRepo)

	s := testService(t, str)
	resp, err := s.GetMe(context.Background(), TestUser1.ID)
	if assert.NotNil(t, resp) {
		expected := &model.GetMeResponse{
			ID:     TestUser1.ID,
			Email:  TestUser1.Email,
			Groups: []model.GroupInUser{},
		}
		assert.Equal(t, expected, resp)
	}
	assert.NoError(t, err)
}

func TestService_GetMe_ErrWhileGetTasksByGroup_Internal(t *testing.T) {
	ctrl := gomock.NewController(t)
	str := mocks.NewMockStore(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	groupRepo := mocks.NewMockGroupRepository(ctrl)
	taskRepo := mocks.NewMockTaskRepository(ctrl)

	userRepo.EXPECT().Get(gomock.Any(), TestUser1.ID).Return(TestUser1, nil)
	groupRepo.EXPECT().GetByUser(gomock.Any(), TestUser1.ID).Return([]*model.Group{TestGroup1}, nil)
	taskRepo.EXPECT().AllByGroupAndUser(gomock.Any(), TestGroup1.ID, TestUser1.ID).Return(nil, errors.New(""))

	str.EXPECT().User().Return(userRepo)
	str.EXPECT().Group().Return(groupRepo)
	str.EXPECT().Task().Return(taskRepo)

	s := testService(t, str)
	resp, err := s.GetMe(context.Background(), TestUser1.ID)
	assert.Nil(t, resp)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, service.ErrInternal)
	}
}

func TestService_GetMe_ErrWhileGetTasksByGroup_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	str := mocks.NewMockStore(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	groupRepo := mocks.NewMockGroupRepository(ctrl)
	taskRepo := mocks.NewMockTaskRepository(ctrl)

	userRepo.EXPECT().Get(gomock.Any(), TestUser1.ID).Return(TestUser1, nil)
	groupRepo.EXPECT().GetByUser(gomock.Any(), TestUser1.ID).Return([]*model.Group{TestGroup1}, nil)
	taskRepo.EXPECT().AllByGroupAndUser(gomock.Any(), TestGroup1.ID, TestUser1.ID).Return(nil, store.ErrNotFound)

	str.EXPECT().User().Return(userRepo)
	str.EXPECT().Group().Return(groupRepo)
	str.EXPECT().Task().Return(taskRepo)

	s := testService(t, str)
	resp, err := s.GetMe(context.Background(), TestUser1.ID)
	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		expected := &model.GetMeResponse{
			ID:    TestUser1.ID,
			Email: TestUser1.Email,
			Groups: []model.GroupInUser{
				{
					ID:          TestGroup1.ID,
					Name:        TestGroup1.Name,
					Description: TestGroup1.Description,
					Tasks:       nil,
				},
			},
		}
		assert.Equal(t, expected, resp)
	}
}

func TestService_GetMe_Positive(t *testing.T) {
	ctrl := gomock.NewController(t)
	str := mocks.NewMockStore(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	groupRepo := mocks.NewMockGroupRepository(ctrl)
	taskRepo := mocks.NewMockTaskRepository(ctrl)

	userRepo.EXPECT().Get(gomock.Any(), TestUser1.ID).Return(TestUser1, nil)
	groupRepo.EXPECT().GetByUser(gomock.Any(), TestUser1.ID).Return([]*model.Group{TestGroup1}, nil)
	taskRepo.EXPECT().AllByGroupAndUser(gomock.Any(), TestGroup1.ID, TestUser1.ID).Return([]*model.Task{TestTask1}, nil)

	str.EXPECT().User().Return(userRepo)
	str.EXPECT().Group().Return(groupRepo)
	str.EXPECT().Task().Return(taskRepo)

	s := testService(t, str)
	resp, err := s.GetMe(context.Background(), TestUser1.ID)
	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		expected := &model.GetMeResponse{
			ID:    TestUser1.ID,
			Email: TestUser1.Email,
			Groups: []model.GroupInUser{
				{
					ID:          TestGroup1.ID,
					Name:        TestGroup1.Name,
					Description: TestGroup1.Description,
					Tasks:       []*model.Task{TestTask1},
				},
			},
		}
		assert.Equal(t, expected, resp)
	}
}
