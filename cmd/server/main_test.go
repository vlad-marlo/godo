package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
)

func TestValidateApp(t *testing.T) {
	assert.NoError(t, fx.ValidateApp(CreateApp()))
}
