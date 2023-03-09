package production

import (
	"github.com/stretchr/testify/assert"
	"github.com/vlad-marlo/godo/internal/config"
	"testing"
)

func TestService_getKeyFunc(t *testing.T) {
	s := testService(t, nil)
	b, err := s.jwtKeyFunc(nil)
	assert.NoError(t, err)
	assert.Equal(t, []byte(config.New().Server.SecretKey), b)
}
