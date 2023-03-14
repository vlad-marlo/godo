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

// CreateGroup creates group in storage and prepares response to user.
func (s *Service) CreateGroup(ctx context.Context, user uuid.UUID, name, description string) (*model.CreateGroupResponse, error) {
	if user == uuid.Nil {
		return nil, service.ErrBadAuthCredentials
	}
	grp := &model.Group{
		ID:          uuid.New(),
		Name:        name,
		Owner:       user,
		Description: description,
	}

	if err := s.store.Group().Create(ctx, grp); err != nil {
		switch {
		case errors.Is(err, store.ErrUniqueViolation):
			return nil, service.ErrGroupAlreadyExists
		case errors.Is(err, store.ErrFKViolation):
			return nil, service.ErrBadData
		default:
		}

		return nil, service.ErrInternal.With(zap.Error(err))
	}
	// TODO: give admin role to user;

	return &model.CreateGroupResponse{
		ID:          grp.ID,
		Name:        name,
		Description: description,
		CreatedAt:   grp.CreatedAt.Unix(),
	}, nil
}

// UseInvite check
func (s *Service) UseInvite(ctx context.Context, user uuid.UUID, group uuid.UUID, invite uuid.UUID) error {
	if !s.store.Invite().Exists(ctx, invite, group) {
		return service.ErrBadInvite
	}

	if err := s.store.Invite().Use(ctx, invite, user); err != nil {
		switch {

		case errors.Is(err, store.ErrInviteIsAlreadyUsed):
			return service.ErrAlreadyInGroup.With(zap.Error(err))

		case errors.Is(err, store.ErrBadData):
			return service.ErrBadInvite.With(zap.Error(err))

		case errors.Is(err, store.ErrUnknown):
			return service.ErrInternal.With(zap.Error(err))

		default:
			return service.ErrInternal.With(zap.Error(err), zap.Stack("stack-trace"))
		}
	}

	return nil
}
