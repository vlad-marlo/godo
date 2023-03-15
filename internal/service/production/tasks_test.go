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
	"time"
)

func TestService_CreateTask_Positive(t *testing.T) {
	tt := []struct {
		name   string
		err    error
		expect error
	}{
		{"unknown", store.ErrUnknown, service.ErrInternal},
		{"FK violation", store.ErrFKViolation, service.ErrBadData},
		{"unique violation", store.ErrUniqueViolation, service.ErrTaskAlreadyExists},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			str := mocks.NewMockStore(ctrl)
			grpID := uuid.New()

			req := model.TaskCreateRequest{
				Name:        uuid.NewString(),
				Description: uuid.NewString(),
				Users:       []uuid.UUID{uuid.New(), uuid.New(), uuid.New()},
				Group:       &grpID,
			}

			taskRepo := mocks.NewMockTaskRepository(ctrl)

			taskRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(tc.err)
			str.EXPECT().Task().Return(taskRepo)

			s := testService(t, str)

			resp, err := s.CreateTask(context.Background(), uuid.Nil, req)
			assert.Nil(t, resp)
			if assert.Error(t, err) {
				assert.ErrorIs(t, err, tc.expect)
			}
		})
	}
}

func TestService_CreateTask_Negative(t *testing.T) {
	tt := []struct {
		name string
		err  error
		want error
	}{
		{"unknown", errors.New(""), service.ErrInternal},
		{"FK violation", store.ErrFKViolation, service.ErrBadData},
		{"unique violation", store.ErrUniqueViolation, service.ErrTaskAlreadyExists},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			taskRepo := mocks.NewMockTaskRepository(ctrl)
			taskRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(tc.err)
			str := mocks.NewMockStore(ctrl)
			str.EXPECT().Task().Return(taskRepo)

			s := testService(t, str)
			task, err := s.CreateTask(context.Background(), uuid.Nil, model.TaskCreateRequest{})
			assert.Nil(t, task)
			if assert.Error(t, err) {
				assert.ErrorIs(t, err, tc.want)
			}
		})
	}
}

func TestService_GetUserTasks_Positive(t *testing.T) {
	ctrl := gomock.NewController(t)
	taskRepo := mocks.NewMockTaskRepository(ctrl)
	tasks := []*model.Task{
		{uuid.New(), uuid.NewString(), uuid.NewString(), time.Now(), uuid.Nil, "NEW"},
		{uuid.New(), uuid.NewString(), uuid.NewString(), time.Now(), uuid.Nil, uuid.NewString()},
		{uuid.New(), uuid.NewString(), uuid.NewString(), time.Now(), uuid.Nil, uuid.NewString()},
		{uuid.New(), uuid.NewString(), uuid.NewString(), time.Now(), uuid.Nil, uuid.NewString()},
		{uuid.New(), uuid.NewString(), uuid.NewString(), time.Now(), uuid.Nil, uuid.NewString()},
	}
	taskRepo.EXPECT().AllByUser(gomock.Any(), uuid.Nil).Return(tasks, nil)
	str := mocks.NewMockStore(ctrl)
	str.EXPECT().Task().Return(taskRepo)

	s := testService(t, str)
	resp, err := s.GetUserTasks(context.Background(), uuid.Nil)
	require.NoError(t, err)
	expected := &model.GetTasksResponse{
		Count: len(tasks),
		Tasks: tasks,
	}
	assert.Equal(t, expected, resp)
}

func TestService_GetUserTasks_Negative(t *testing.T) {
	tt := []struct {
		name   string
		err    error
		expect error
	}{
		{"unknown", errors.New(""), service.ErrInternal},
		{"not found", store.ErrNotFound, service.ErrNoContent},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			taskRepo := mocks.NewMockTaskRepository(ctrl)
			taskRepo.EXPECT().AllByUser(gomock.Any(), uuid.Nil).Return(nil, tc.err)
			str := mocks.NewMockStore(ctrl)
			str.EXPECT().Task().Return(taskRepo)

			s := testService(t, str)
			resp, err := s.GetUserTasks(context.Background(), uuid.Nil)
			require.Nil(t, resp)
			if assert.Error(t, err) {
				assert.ErrorIs(t, err, tc.expect)
			}
		})
	}
}

func TestService_GetTask_Positive(t *testing.T) {
	ctrl := gomock.NewController(t)
	taskRepo := mocks.NewMockTaskRepository(ctrl)
	task := &model.Task{
		ID:          uuid.New(),
		Name:        uuid.NewString(),
		Description: uuid.NewString(),
		CreatedAt:   time.Now(),
		CreatedBy:   uuid.Nil,
		Status:      uuid.NewString(),
	}
	taskRepo.EXPECT().GetByUserAndID(gomock.Any(), uuid.Nil, uuid.Nil).Return(task, nil)
	st := mocks.NewMockStore(ctrl)
	st.EXPECT().Task().Return(taskRepo)

	s := testService(t, st)
	got, err := s.GetTask(context.Background(), uuid.Nil, uuid.Nil)
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.Equal(t, task, got)
	}
}

func TestService_GetTask_Negative(t *testing.T) {
	tt := []struct {
		name string
		err  error
		want error
	}{
		{"unknown", errors.New(""), service.ErrInternal},
		{"not found", store.ErrNotFound, service.ErrNotFound},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			taskRepo := mocks.NewMockTaskRepository(ctrl)
			taskRepo.EXPECT().GetByUserAndID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, tc.err)
			st := mocks.NewMockStore(ctrl)
			st.EXPECT().Task().Return(taskRepo)

			s := testService(t, st)
			got, err := s.GetTask(context.Background(), uuid.New(), uuid.New())
			assert.Nil(t, got)
			if assert.Error(t, err) {
				assert.ErrorIs(t, err, tc.want)
			}
		})
	}
}

func TestService_AddTaskToUser(t *testing.T) {
	tt := []struct {
		name string
		err  error
	}{
		{"nil", nil},
		{"not nil", errors.New("")},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			taskRepo := mocks.NewMockTaskRepository(ctrl)
			taskRepo.EXPECT().AddToUser(gomock.Any(), uuid.Nil, uuid.Nil, uuid.Nil).Return(tc.err)
			str := mocks.NewMockStore(ctrl)
			str.EXPECT().Task().Return(taskRepo)
			s := testService(t, str)
			s.addTaskToUser(context.Background(), uuid.Nil, uuid.Nil, uuid.Nil, 1)
		})
	}

}

func TestService_AddTaskToUsers(t *testing.T) {

}

func TestService_AddTaskToGroup_Negative_NoPermission(t *testing.T) {
	role := &model.Role{
		ID:       0,
		Members:  model.PermCreate,
		Tasks:    model.PermReadRelated,
		Reviews:  model.PermChangeRelated,
		Comments: model.PermChangeAll,
	}
	ctrl := gomock.NewController(t)
	groupRepo := mocks.NewMockGroupRepository(ctrl)
	groupRepo.EXPECT().GetRoleOfMember(gomock.Any(), uuid.Nil, uuid.Nil).Return(role, nil)
	str := mocks.NewMockStore(ctrl)
	str.EXPECT().Group().Return(groupRepo)
	s := testService(t, str)
	s.addTaskToGroup(context.Background(), uuid.Nil, uuid.Nil, uuid.Nil)
}

func TestService_AddTaskToGroup_Positive(t *testing.T) {
	ctrl := gomock.NewController(t)

	role := &model.Role{
		ID:       0,
		Members:  model.PermCreate,
		Tasks:    model.PermChangeAll,
		Reviews:  model.PermChangeRelated,
		Comments: model.PermChangeAll,
	}

	str := mocks.NewMockStore(ctrl)
	taskRepo := mocks.NewMockTaskRepository(ctrl)
	groupRepo := mocks.NewMockGroupRepository(ctrl)

	groupRepo.EXPECT().GetRoleOfMember(gomock.Any(), uuid.Nil, uuid.Nil).Return(role, nil)
	taskRepo.EXPECT().AddToGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	str.EXPECT().Group().Return(groupRepo)
	str.EXPECT().Task().Return(taskRepo)

	s := testService(t, str)
	s.addTaskToGroup(context.Background(), uuid.Nil, uuid.Nil, uuid.Nil)
}

func TestService_AddTaskToGroup_Negative_ErrWhileAdding(t *testing.T) {
	role := &model.Role{
		ID:       0,
		Members:  model.PermCreate,
		Tasks:    model.PermChangeAll,
		Reviews:  model.PermChangeRelated,
		Comments: model.PermChangeAll,
	}
	ctrl := gomock.NewController(t)
	groupRepo := mocks.NewMockGroupRepository(ctrl)
	groupRepo.EXPECT().GetRoleOfMember(gomock.Any(), uuid.Nil, uuid.Nil).Return(role, nil)
	taskRepo := mocks.NewMockTaskRepository(ctrl)
	taskRepo.EXPECT().AddToGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New(""))
	str := mocks.NewMockStore(ctrl)
	str.EXPECT().Group().Return(groupRepo)
	str.EXPECT().Task().Return(taskRepo)
	s := testService(t, str)
	s.addTaskToGroup(context.Background(), uuid.Nil, uuid.Nil, uuid.Nil)
}

func TestService_AddTaskToGroup_Negative_ErrWhileGettingRole(t *testing.T) {
	ctrl := gomock.NewController(t)

	str := mocks.NewMockStore(ctrl)
	groupRepo := mocks.NewMockGroupRepository(ctrl)

	groupRepo.EXPECT().GetRoleOfMember(gomock.Any(), uuid.Nil, uuid.Nil).Return(nil, errors.New(""))

	str.EXPECT().Group().Return(groupRepo)

	s := testService(t, str)
	s.addTaskToGroup(context.Background(), uuid.Nil, uuid.Nil, uuid.Nil)
}
