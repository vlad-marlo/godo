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

// GetUserTasks return tasks related to user with provided id.
func (s *Service) GetUserTasks(ctx context.Context, user uuid.UUID) (*model.GetTasksResponse, error) {
	tasks, err := s.store.Task().AllByUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			return nil, service.ErrNoContent.With(zap.Error(err))
		default:
			return nil, service.ErrInternal.With(zap.Error(err))
		}
	}

	return &model.GetTasksResponse{
		Count: len(tasks),
		Tasks: tasks,
	}, nil
}

// GetTask return task by user and task id.
func (s *Service) GetTask(ctx context.Context, user, task uuid.UUID) (*model.Task, error) {
	t, err := s.store.Task().GetByUserAndID(ctx, user, task)
	if err != nil {
		switch {

		case errors.Is(err, store.ErrNotFound):
			return nil, service.ErrNotFound

		default:
			return nil, service.ErrInternal.With(zap.Error(err))

		}
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
	if r.Tasks >= model.PermCreate {
		if err = s.store.Task().AddToGroup(ctx, task, user); err != nil {
			s.log.Warn("error while adding task to group", zap.Error(err))
		}
	}
}

func (s *Service) addToUsers(ctx context.Context, user, task uuid.UUID, users []uuid.UUID) {
	for p, u := range users {
		select {
		case <-ctx.Done():
			return
		default:
			go s.addTaskToUser(ctx, user, task, u, p)
		}
	}
}

func (s *Service) addToGroupUsers(ctx context.Context, group, task uuid.UUID) {
	ids, err := s.store.Group().GetUserIDs(ctx, group)
	if err != nil {
		s.log.Error("service: add to group users: store: group: get user ids", zap.Error(err))
		return
	}

	for p, u := range ids {
		select {
		case <-ctx.Done():
			return
		default:
			if err = s.store.Task().ForceAddToUser(ctx, u, task); err != nil {
				s.log.Error("add task to user", zap.Error(err), zap.Int("pool", p))
			}
		}
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
		Status:      "NEW",
	}

	if err := s.store.Task().Create(ctx, task); err != nil {
		switch {
		case errors.Is(err, store.ErrUniqueViolation):
			return nil, service.ErrTaskAlreadyExists

		case errors.Is(err, store.ErrFKViolation):
			return nil, service.ErrBadData

		default:
			return nil, service.ErrInternal.With(zap.Error(err))
		}
	}

	// async add task to group and users
	if req.Group != nil {
		go s.addTaskToGroup(context.Background(), user, task.ID, *req.Group)
		if req.Users == nil {
			go s.addToGroupUsers(context.Background(), *req.Group, task.ID)
		}
	}
	if req.Users != nil {
		go s.addToUsers(context.Background(), user, task.ID, req.Users)
	}

	return task, nil
}
