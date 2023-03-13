package production

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/service"
	"github.com/vlad-marlo/godo/internal/store"
	"github.com/vlad-marlo/godo/internal/store/mocks"
	"testing"
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
