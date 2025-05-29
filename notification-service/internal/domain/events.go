package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRegisteredEvent struct {
	UserID       uuid.UUID `json:"user_id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	RegisteredAt time.Time `json:"registered_at"`
}
