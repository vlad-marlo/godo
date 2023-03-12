package production

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/service"
	"github.com/vlad-marlo/godo/internal/store"
	"go.uber.org/zap"
	"time"
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

func (s *Service) addTaskToUser(ctx context.Context, user, task, to uuid.UUID, pool int) {
	if err := s.store.Task().AddToUser(ctx, user, task, to); err != nil {
		s.log.Warn("error while adding task to user", zap.Int("pool", pool))
	}
}

func (s *Service) addTaskToGroup(ctx context.Context, user, task, group uuid.UUID) {
	r, err := s.store.Group().GetRoleOfMember(ctx, user, group)
	if err != nil {
		s.log.Warn("error while getting role of member", zap.Error(err))
		return
	}
	if r.Tasks >= 2 {
		if err = s.store.Task().AddToGroup(ctx, task, user); err != nil {
			s.log.Warn("error while adding task to group", zap.Error(err))
		}
	}
}

func (s *Service) addToUsers(ctx context.Context, user, task uuid.UUID, users []uuid.UUID) {
	for p, u := range users {
		go s.addTaskToUser(ctx, user, task, u, p)
	}
}

// CreateTask create record about task in database.
func (s *Service) CreateTask(ctx context.Context, user uuid.UUID, req model.TaskCreateRequest) (*model.Task, error) {
	task := &model.Task{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
		CreatedBy:   user,
		Created:     0,
		Status:      "NEW",
	}
	if err := s.store.Task().Create(ctx, task); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, service.ErrNotFound
		}

		return nil, service.ErrInternal.With(zap.Error(err))
	}

	// async add task to group and users
	go s.addTaskToGroup(ctx, user, task.ID, req.Group)
	go s.addToUsers(ctx, user, task.ID, req.Users)

	return task, nil
}
