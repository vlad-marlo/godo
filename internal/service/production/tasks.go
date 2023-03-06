package production

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/service"
	"github.com/vlad-marlo/godo/internal/store"
	"go.uber.org/zap"
)

func (s *Service) GetUserTasks(ctx context.Context, user uuid.UUID) (*model.GetTasksResponse, error) {
	tasks, err := s.store.Task().AllByUser(ctx, user)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, service.ErrNotFound.With(zap.Error(err))
		}
		return nil, service.ErrInternal.With(zap.Error(err))
	}

	return &model.GetTasksResponse{
		Count: len(tasks),
		Tasks: tasks,
	}, nil
}

func (s *Service) GetTask(ctx context.Context, user, task uuid.UUID) (*model.Task, error) {
	t, err := s.store.Task().GetByUserAndID(ctx, user, task)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, service.ErrNotFound
		}
		return nil, service.ErrInternal.With(zap.Error(err))
	}
	return t, nil
}
