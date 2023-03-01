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
		return nil, service.ErrBadUser.With(
			zap.String("user", user.String()),
		)
	}
	grp := &model.Group{
		ID:          uuid.New(),
		Name:        name,
		Owner:       user,
		Description: description,
	}

	if err := s.store.Group().Create(ctx, grp); err != nil {
		if errors.Is(err, store.ErrGroupAlreadyExists) {
			return nil, service.ErrGroupAlreadyExists
		}

		return nil, service.ErrInternal.With(zap.Error(err))
	}

	return &model.CreateGroupResponse{
		ID:          grp.ID,
		Name:        name,
		Description: description,
		CreatedAt:   grp.CreatedAt.Unix(),
	}, nil
}

func (s *Service) AddUserToGroup() (*model.InviteUserInGroupResponse, error) {
	//TODO: implement me.
	panic("not implemented")
}
