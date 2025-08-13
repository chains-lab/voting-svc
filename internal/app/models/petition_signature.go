package models

import (
	"time"

	"github.com/google/uuid"
)

type PetitionSignature struct {
	ID         uuid.UUID
	PetitionID uuid.UUID
	UserID     uuid.UUID
	CreatedAt  time.Time
}
