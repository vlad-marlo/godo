package model

import (
	"github.com/google/uuid"
	"time"
)

type Token struct {
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	Expires   bool
}
