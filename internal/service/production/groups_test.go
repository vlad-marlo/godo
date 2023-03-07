package production

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/service"
	"github.com/vlad-marlo/godo/internal/store"
	"github.com/vlad-marlo/godo/internal/store/mocks"
	"testing"
)

func TestService_CreateGroup_Positive(t *testing.T) {
	ctrl := gomock.NewController(t)

	str := mocks.NewMockStore(ctrl)
	grp := mocks.NewMockGroupRepository(ctrl)
	grp.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, g *model.Group) error {
		assert.Equal(t, TestGroup1.Owner, g.Owner)
		return nil
	})

	str.EXPECT().Group().Return(grp)

	srv := testService(t, str)

	ctx := context.Background()

	resp, err := srv.CreateGroup(ctx, TestGroup1.Owner, TestGroup1.Name, TestGroup1.Description)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	assert.Equal(t, TestGroup1.Name, resp.Name)
	assert.Equal(t, TestGroup1.Description, resp.Description)
}

func TestService_CreateGroup_Negative_ErrGroupAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	str := mocks.NewMockStore(ctrl)
	grp := mocks.NewMockGroupRepository(ctrl)
	grp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(store.ErrUniqueViolation)

	str.EXPECT().Group().Return(grp)

	srv := testService(t, str)

	resp, err := srv.CreateGroup(context.Background(), TestGroup1.Owner, TestGroup1.Name, TestGroup1.Description)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, service.ErrGroupAlreadyExists)
}

func TestService_CreateGroup_BadUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	str := mocks.NewMockStore(ctrl)
	srv := testService(t, str)

	resp, err := srv.CreateGroup(context.Background(), uuid.Nil, TestGroup1.Name, TestGroup1.Description)
	assert.ErrorIs(t, err, service.ErrBadAuthCredentials)
	assert.Nil(t, resp)
}

func TestService_CreateGroup_BadRequest(t *testing.T) {
	const errMsg = "error message"
	ctrl := gomock.NewController(t)
	str := mocks.NewMockStore(ctrl)
	grp := mocks.NewMockGroupRepository(ctrl)
	grp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New(errMsg))
	str.EXPECT().Group().Return(grp)

	srv := testService(t, str)
	resp, err := srv.CreateGroup(context.Background(), TestGroup1.Owner, TestGroup1.Name, TestGroup1.Description)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, service.ErrInternal)
}

func TestService_UseInvite(t *testing.T) {
	t.Run("invite does not exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		inv := mocks.NewMockInviteRepository(ctrl)
		inv.EXPECT().Exists(gomock.Any(), gomock.Any(), gomock.Any()).Return(false)

		st := mocks.NewMockStore(ctrl)
		st.EXPECT().Invite().Return(inv)

		srv := testService(t, st)

		err := srv.UseInvite(context.Background(), uuid.New(), uuid.New(), uuid.New())
		require.Error(t, err)
		assert.ErrorIs(t, err, service.ErrBadInvite)
	})
	tt := []struct {
		name string
		err  error
		want error
		ass  assert.ErrorAssertionFunc
	}{
		{"already used", store.ErrInviteIsAlreadyUsed, service.ErrAlreadyInGroup, assert.Error},
		{"bad data", store.ErrBadData, service.ErrBadInvite, assert.Error},
		{"unknown store", store.ErrUnknown, service.ErrInternal, assert.Error},
		{"unknown really unknown", errors.New(""), service.ErrInternal, assert.Error},
		{"nil", nil, nil, assert.NoError},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			inv := mocks.NewMockInviteRepository(ctrl)
			inv.EXPECT().Exists(gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

			inv.EXPECT().Use(gomock.Any(), gomock.Any(), gomock.Any()).Return(tc.err).AnyTimes()
			st := mocks.NewMockStore(ctrl)
			st.EXPECT().Invite().Return(inv).AnyTimes()

			srv := testService(t, st)

			err := srv.UseInvite(context.Background(), uuid.New(), uuid.New(), uuid.New())
			tc.ass(t, err)
			assert.ErrorIs(t, err, tc.want)
		})
	}
}
