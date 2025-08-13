package models

import (
	"time"

	"github.com/google/uuid"
)

type PollVote struct {
	ID        uuid.UUID
	PollID    uuid.UUID
	UserID    uuid.UUID
	OptionID  uuid.UUID
	CreatedAt time.Time
}
