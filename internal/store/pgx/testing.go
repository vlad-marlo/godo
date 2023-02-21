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

var _dbTables = []string{"users", "groups"}

var (
	TestUser1 = &model.User{
		ID:   uuid.New(),
		Name: "first_user",
		Pass: "second_password",
	}
	TestUser2 = &model.User{
		ID:   uuid.New(),
		Name: "second_user",
		Pass: "some_passwords",
	}
	TestUser3 = &model.User{
		ID:   uuid.New(),
		Name: "first_user",
		Pass: "some_password",
	}
	TestGroup1 = &model.Group{
		ID:          uuid.New(),
		Name:        "test group",
		CreatedBy:   TestUser1.ID,
		Description: "test description",
		CreatedAt:   time.Now(),
	}
	TestGroup2 = &model.Group{
		ID:          uuid.New(),
		Name:        "another test group",
		CreatedBy:   TestUser1.ID,
		Description: "another description",
		CreatedAt:   time.Now(),
	}
)

// testStore ...
func testStore(t testing.TB, cli Client) (*Store, func()) {
	t.Helper()
	if cli == nil {
		cli = postgres.TestClient(t)
	}
	u := NewUserRepository(cli)
	g := NewGroupRepository(cli)
	s := New(cli, u, g)
	return s, func() {
		defer s.Close()
		_, err := s.p.Exec(context.Background(), fmt.Sprintf(`TRUNCATE %s CASCADE;`, strings.Join(_dbTables, ", ")))
		assert.NoError(t, err)
	}
}

// testUsers ...
func testUsers(t testing.TB) (*UserRepository, func()) {
	t.Helper()
	cli := postgres.TestClient(t)
	s := NewUserRepository(cli)

	return s, func() {
		defer cli.Close()
		_, err := s.p.Exec(context.Background(), fmt.Sprintf(`TRUNCATE users CASCADE;`))
		assert.NoError(t, err)
	}
}

// testGroup ...
func testGroup(t testing.TB) (*GroupRepository, func()) {
	t.Helper()
	cli := postgres.TestClient(t)
	s := NewGroupRepository(cli)

	return s, func() {
		defer cli.Close()
		_, err := s.pool.Exec(context.Background(), fmt.Sprintf(`TRUNCATE groups CASCADE;`))
		assert.NoError(t, err)
	}
}

// testGroupUser ...
func testGroupUser(t testing.TB) (*GroupRepository, *UserRepository, func()) {
	t.Helper()
	cli := postgres.TestClient(t)
	grp := NewGroupRepository(cli)
	usr := NewUserRepository(cli)

	return grp, usr, func() {
		defer cli.Close()
		_, err := cli.P().Exec(context.Background(), fmt.Sprintf(`TRUNCATE groups, users CASCADE;`))
		assert.NoError(t, err)
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
