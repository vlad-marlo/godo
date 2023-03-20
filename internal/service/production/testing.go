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
	TestTask1 = &model.Task{
		ID:          uuid.New(),
		Name:        uuid.NewString(),
		Description: uuid.NewString(),
		CreatedAt:   time.Now(),
		CreatedBy:   TestUser1.ID,
		Status:      uuid.NewString(),
	}
	ReadOnlyRole = &model.Role{
		ID:       0,
		Members:  model.PermReadRelated,
		Tasks:    model.PermReadRelated,
		Reviews:  model.PermReadRelated,
		Comments: model.PermReadRelated,
	}
	SudoRole = &model.Role{
		ID:       0,
		Members:  model.PermChangeAll,
		Tasks:    model.PermChangeAll,
		Reviews:  model.PermChangeAll,
		Comments: model.PermChangeAll,
	}
)

func testService(t testing.TB, s store.Store) *Service {
	t.Helper()
	return New(s, config.New(), zap.L())
}
