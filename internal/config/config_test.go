package config

import (
	"github.com/stretchr/testify/assert"
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
