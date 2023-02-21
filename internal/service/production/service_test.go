package production

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"log"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/vlad-marlo/godo/internal/config"
	"github.com/vlad-marlo/godo/internal/store/mocks"
)

func TestService_Ping_Positive(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mocks.NewMockStore(ctrl)
	store.EXPECT().Ping(gomock.Any()).Return(nil)
	s := &Service{store: store}
	assert.NoError(t, s.Ping(context.Background()))
}

func TestService_Ping_Negative(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mocks.NewMockStore(ctrl)
	store.EXPECT().Ping(gomock.Any()).Return(errors.New(""))
	s := &Service{store: store}
	assert.Error(t, s.Ping(context.Background()))
}

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mocks.NewMockStore(ctrl)
	s := New(store, config.New(), zap.L())
	assert.NotNil(t, s)
}

func TestMain(m *testing.M) {
	if err := os.Setenv("TEST", "true"); err != nil {
		log.Fatalf("got unexpected error while setting up env: %v", err)
	}
	os.Exit(m.Run())
}
