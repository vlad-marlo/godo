package model

import (
	"time"

	"github.com/google/uuid"
)

// Token is access token struct.
type Token struct {
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	Expires   bool
}
