package pgx

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/vlad-marlo/godo/internal/store/pgx/mocks"
	"go.uber.org/zap"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/pkg/client/postgres"
)

var _dbTables = []string{"users", "groups", "auth_tokens", "comments", "reviews", "roles", "task_group", "task_user", "tasks", "user_in_group", "invites"}

var (
	TestUser1 = &model.User{
		ID:    uuid.New(),
		Pass:  "second_password",
		Email: "testemail1@xd.ru",
	}
	TestUser2 = &model.User{
		ID:    uuid.New(),
		Pass:  "some_passwords",
		Email: "good_email2@example.com",
	}
	TestUser3 = &model.User{
		ID:    uuid.New(),
		Pass:  "some_password",
		Email: TestUser1.Email,
	}
	TestGroup1 = &model.Group{
		ID:          uuid.New(),
		Name:        "test group",
		Owner:       TestUser1.ID,
		Description: "test description",
		CreatedAt:   time.Now(),
	}
	TestGroup2 = &model.Group{
		ID:          uuid.New(),
		Name:        "another test group",
		Owner:       TestUser1.ID,
		Description: "another description",
		CreatedAt:   time.Now(),
	}
	TestToken1 = &model.Token{
		UserID:    TestUser1.ID,
		Token:     "some token",
		ExpiresAt: time.Now().UTC(),
		Expires:   true,
	}
	//TestToken2 = &model.Token{
	//	UserID:    TestUser1.ID,
	//	Token:     "another token",
	//	ExpiresAt: time.Now().Add(time.Hour),
	//	Expires:   true,
	//}
	//TestToken3 = &model.Token{
	//	UserID:    TestUser1.ID,
	//	Token:     TestToken1.Token,
	//	ExpiresAt: time.Now(),
	//	Expires:   false,
	//}
)

// testStore ...
func testStore(t testing.TB, cli Client) (*Store, func()) {
	t.Helper()
	if cli == nil {
		cli = postgres.TestClient(t)
	}
	userRepo := NewUserRepository(cli)
	groupRepo := NewGroupRepository(cli)
	tokenRepo := NewTokenRepository(cli)
	s := New(cli, userRepo, groupRepo, tokenRepo)
	return s, func() {
		teardown(t, cli)(_dbTables...)
	}
}

// testUsers ...
func testUsers(t testing.TB) (*UserRepository, func()) {
	t.Helper()
	cli := postgres.TestClient(t)
	s := NewUserRepository(cli)

	return s, func() {
		teardown(t, cli)("users")
	}
}

// testGroup ...
func testGroup(t testing.TB) (*GroupRepository, func()) {
	t.Helper()
	cli := postgres.TestClient(t)
	s := NewGroupRepository(cli)

	return s, func() {
		teardown(t, cli)("groups")
	}
}

// testGroupUser ...
func testGroupUser(t testing.TB) (*GroupRepository, *UserRepository, func()) {
	t.Helper()
	cli := postgres.TestClient(t)
	grp := NewGroupRepository(cli)
	usr := NewUserRepository(cli)

	return grp, usr, func() {
		teardown(t, cli)("users", "groups", "user_in_group")
	}
}

func teardown(t testing.TB, cli Client) func(...string) {
	return func(tables ...string) {
		closer, ok := cli.(interface{ Close() })
		if ok {
			defer closer.Close()
		}

		if len(tables) > 1 {
			_, err := cli.P().Exec(context.Background(), fmt.Sprintf(`TRUNCATE %s CASCADE;`, strings.Join(tables, ", ")))
			assert.NoError(t, err)
		} else if len(tables) == 1 {
			_, err := cli.P().Exec(context.Background(), fmt.Sprintf(`TRUNCATE %s CASCADE;`, tables[0]))
			assert.NoError(t, err)
		}
	}
}

func BadCli(t testing.TB) Client {
	ctrl := gomock.NewController(t)
	cli := mocks.NewMockClient(ctrl)
	cli.EXPECT().L().Return(zap.L()).AnyTimes()

	pool, err := pgxpool.New(context.Background(), "postgresql://a:a@a:1/a")
	require.NoError(t, err)

	cli.EXPECT().P().Return(pool).AnyTimes()
	return cli
}
