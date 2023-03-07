package pgx

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vlad-marlo/godo/internal/store/mocks"
	"testing"
)

func TestMockExists(t *testing.T) {
	tt := []bool{true, false}
	for _, exp := range tt {
		t.Run("exp", func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			userRepo := mocks.NewMockUserRepository(ctrl)
			invRepo := mocks.NewMockInviteRepository(ctrl)
			userRepo.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(exp)
			invRepo.EXPECT().Exists(gomock.Any(), gomock.Any(), gomock.Any()).Return(exp)

			assert.Equal(t, exp, userRepo.Exists(ctx, ""))
			assert.Equal(t, exp, invRepo.Exists(ctx, uuid.Nil, uuid.Nil))
		})
	}

}
