package store

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vlad-marlo/godo/internal/store/mocks"
	"testing"
)

func TestMockExists(t *testing.T) {
	tt := []bool{true, false}
	for _, exp := range tt {
		t.Run("exp", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s := mocks.NewMockStore(ctrl)
			u := mocks.NewMockUserRepository(ctrl)
			g := mocks.NewMockGroupRepository(ctrl)
			u.EXPECT().
				Exists(gomock.Any(), gomock.Any()).
				Return(exp)
			g.EXPECT().
				Exists(gomock.Any(), gomock.Any()).
				Return(exp)
			s.EXPECT().
				User().
				Return(u)
			s.EXPECT().
				Group().
				Return(g)

			assert.Equal(t, exp, u.Exists(context.Background(), ""))
			assert.Equal(t, exp, g.Exists(context.Background(), ""))
		})
	}

}
