package models

import (
	"time"

	"github.com/google/uuid"
)

type PollOption struct {
	ID         uuid.UUID
	PollID     uuid.UUID
	OptionText string
	VotesCount int
	CreatedAt  time.Time
}
