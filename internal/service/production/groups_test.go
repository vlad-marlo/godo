package production

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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

	resp, err := srv.CreateGroup(ctx, TestGroup1.Owner.String(), TestGroup1.Name, TestGroup1.Description)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	assert.Equal(t, TestGroup1.Name, resp.Name)
	assert.Equal(t, TestGroup1.Description, resp.Description)
}

func TestService_CreateGroup_Negative_ErrGroupAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	str := mocks.NewMockStore(ctrl)
	grp := mocks.NewMockGroupRepository(ctrl)
	grp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(store.ErrGroupAlreadyExists)

	str.EXPECT().Group().Return(grp)

	srv := testService(t, str)

	resp, err := srv.CreateGroup(context.Background(), TestGroup1.Owner.String(), TestGroup1.Name, TestGroup1.Description)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, service.ErrGroupAlreadyExists)
}

func TestService_CreateGroup_BadUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	str := mocks.NewMockStore(ctrl)
	srv := testService(t, str)

	resp, err := srv.CreateGroup(context.Background(), "sdaf", TestGroup1.Name, TestGroup1.Description)
	assert.ErrorIs(t, err, service.ErrInternal)
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
	resp, err := srv.CreateGroup(context.Background(), TestGroup1.Owner.String(), TestGroup1.Name, TestGroup1.Description)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, service.ErrInternal)
}
