package config

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfig_Valid(t *testing.T) {
	cfg := &Config{}
	ok, err := cfg.Valid()
	assert.Error(t, err)
	assert.False(t, ok)
}

func TestGenerateRandom(t *testing.T) {
	b, err := generateRandom(0)
	assert.NoError(t, err)
	assert.Empty(t, b)
}

func Test_byteToString(t *testing.T) {
	b, err := generateRandom(12)
	require.NoError(t, err)
	assert.Equal(t, base64.StdEncoding.EncodeToString(b), byteToString(b))
}
