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
func (s *Service) CreateGroup(ctx context.Context, user, name, description string) (*model.CreateGroupResponse, error) {
	userID, err := uuid.Parse(user)
	if err != nil {
		return nil, service.ErrInternal.With(
			zap.String("user", user),
			zap.Error(err),
		)
	}
	grp := &model.Group{
		ID:          uuid.New(),
		Name:        name,
		Owner:       userID,
		Description: description,
	}

	if err = s.store.Group().Create(ctx, grp); err != nil {
		if errors.Is(err, store.ErrGroupAlreadyExists) {
			return nil, service.ErrGroupAlreadyExists
		}

		return nil, service.ErrInternal.With(zap.Error(err))
	}
	return &model.CreateGroupResponse{
		ID:          grp.ID,
		Name:        name,
		Description: description,
		CreatedAt:   grp.CreatedAt.Format(s.cfg.Server.TimeFormat),
	}, nil
}
