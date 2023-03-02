package pgx

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/vlad-marlo/godo/internal/pkg/client/postgres"
	"github.com/vlad-marlo/godo/internal/store/pgx/mocks"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	cli := mocks.NewMockClient(ctrl)
	p := postgres.TestClient(t).P()
	l := zap.L()
	cli.EXPECT().P().Return(p).AnyTimes()
	cli.EXPECT().L().Return(l).AnyTimes()
	s, td := testStore(t, cli)
	assert.Equal(t, l, s.log)
	assert.Equal(t, p, s.pool)
	defer td()
}

func TestStore_User(t *testing.T) {
	cli := postgres.TestClient(t)
	usrRepo := NewUserRepository(cli)
	grpRepo := NewGroupRepository(cli)
	tokRepo := NewTokenRepository(cli)
	tskRepo := NewTaskRepository(cli)
	invRepo := NewInviteRepository(cli)
	s := New(
		cli,
		usrRepo,
		grpRepo,
		tokRepo,
		tskRepo,
		invRepo,
	)
	assert.Equal(t, usrRepo, s.User())
	assert.Equal(t, s.user, s.User())

	assert.Equal(t, grpRepo, s.Group())
	assert.Equal(t, grpRepo, s.group)

	assert.Equal(t, tokRepo, s.token)
	assert.Equal(t, s.token, s.Token())

	assert.Equal(t, tskRepo, s.Task())
	assert.Equal(t, s.task, s.Task())

	assert.Equal(t, s.invite, s.Invite())
	assert.Equal(t, s.invite, invRepo)
	s.Close()
}

func TestPing(t *testing.T) {
	s, td := testStore(t, nil)
	defer td()
	require.NoError(t, s.Ping(context.Background()))
}

func TestMain(m *testing.M) {
	if err := os.Setenv("TEST", "true"); err != nil {
		log.Fatalf("os: setenv: %s", err.Error())
	}
	os.Exit(m.Run())
}
