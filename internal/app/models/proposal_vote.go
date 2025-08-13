package models

import (
	"time"

	"github.com/google/uuid"
)

type ProposalVote struct {
	ID         uuid.UUID
	ProposalID uuid.UUID
	UserID     uuid.UUID
	Vote       bool
	CreatedAt  time.Time
}
