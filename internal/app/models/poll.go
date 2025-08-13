package models

import (
	"time"

	"github.com/google/uuid"
)

type Poll struct {
	ID          uuid.UUID
	CityID      uuid.UUID
	Title       string
	Description string
	Status      string
	InitiatorID uuid.UUID
	Options     []PollOption
	EndDate     time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Location    *Location
}
