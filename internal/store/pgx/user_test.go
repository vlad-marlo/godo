package pgx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/store"
)

func TestUserRepository_Create(t *testing.T) {
	s, td := testUsers(t)
	defer td()
	tt := []struct {
		name    string
		u       *model.User
		unknown bool
		wantErr error
	}{
		{
			name:    "positive #1",
			u:       TestUser1,
			wantErr: nil,
		},
		{
			name:    "positive #2",
			u:       TestUser2,
			wantErr: nil,
		},
		{
			name:    "negative: duplicates username",
			u:       TestUser3,
			wantErr: store.ErrUserAlreadyExists,
		},
		{
			name:    "negative: nil user",
			u:       nil,
			wantErr: store.ErrBadData,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.unknown {
				assert.Error(t, s.Create(context.Background(), tc.u))
			} else {
				assert.ErrorIs(t, s.Create(context.Background(), tc.u), tc.wantErr)
			}
		})
	}
}

func TestUserRepository_GetByName(t *testing.T) {
	ctx := context.Background()

	s, td := testUsers(t)
	defer td()

	u, err := s.GetByName(ctx, TestUser1.Name)
	assert.Nil(t, u)
	assert.ErrorIs(t, err, store.ErrNotFound)

	assert.False(t, s.Exists(ctx, TestUser1.ID.String()))

	assert.NoError(t, s.Create(ctx, TestUser1))
	u, err = s.GetByName(ctx, TestUser1.Name)

	assert.Equal(t, TestUser1, u)
	assert.NoError(t, err)
	assert.True(t, s.Exists(ctx, TestUser1.ID.String()))

	u, err = s.GetByName(ctx, TestUser2.Name)
	assert.Nil(t, u)
	assert.Error(t, err)
	assert.ErrorIs(t, err, store.ErrNotFound)

	assert.False(t, s.Exists(ctx, TestUser2.ID.String()))
}

func TestUserRepository_GetByName_Negative(t *testing.T) {
	cli := BadCli(t)
	userRepository := NewUserRepository(cli)

	u, err := userRepository.GetByName(context.Background(), "xd")
	assert.Nil(t, u)
	assert.Error(t, err)
	assert.ErrorIs(t, err, store.ErrUnknown)
}

func TestUserRepository_Create_Negative(t *testing.T) {
	cli := BadCli(t)
	userRepository := NewUserRepository(cli)

	err := userRepository.Create(context.Background(), new(model.User))
	assert.Error(t, err)
	assert.ErrorIs(t, err, store.ErrUnknown)
}
