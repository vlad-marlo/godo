package production

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/pkg/fielderr"
	"github.com/vlad-marlo/godo/internal/store"
)

// CreateGroup creates group in storage and prepares response to user.
func (s *Service) CreateGroup(ctx context.Context, user, name, description string) (*model.CreateGroupResponse, error) {
	userID, err := uuid.Parse(user)
	if err != nil {
		return nil, fielderr.New("bad user id", nil, fielderr.CodeUnauthorized)
	}
	grp := &model.Group{
		ID:          uuid.New(),
		Name:        name,
		CreatedBy:   userID,
		Description: description,
	}

	err = s.store.Group().Create(ctx, grp)

	if err != nil {
		if errors.Is(err, store.ErrGroupAlreadyExists) {
			return nil, fielderr.New(err.Error(), nil, fielderr.CodeConflict)
		}

		return nil, fielderr.New(err.Error(), nil, fielderr.CodeBadRequest)
	}
	return &model.CreateGroupResponse{
		ID:          grp.ID,
		Name:        name,
		Description: description,
		CreatedAt:   grp.CreatedAt.Format(s.cfg.Server.TimeFormat),
	}, nil
}
