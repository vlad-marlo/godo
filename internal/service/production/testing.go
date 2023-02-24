package production

import (
	"github.com/google/uuid"
	"github.com/vlad-marlo/godo/internal/config"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/store"
	"go.uber.org/zap"
	"testing"
	"time"
)

var (
	TestUser1 = &model.User{
		ID:    uuid.New(),
		Email: "user1@test.com",
		Pass:  "some pass",
	}
	TestGroup1 = &model.Group{
		ID:          uuid.New(),
		Name:        "test group 1",
		Owner:       TestUser1.ID,
		Description: "test description",
		CreatedAt:   time.Now(),
	}
)

func testService(t testing.TB, s store.Store) *Service {
	t.Helper()
	return New(s, config.New(), zap.L())
}
