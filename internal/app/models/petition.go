package models

import (
	"time"

	"github.com/google/uuid"
)

type Petition struct {
	ID          uuid.UUID
	CityID      uuid.UUID
	Title       string
	Description string
	InitiatorID uuid.UUID
	AddressToID *uuid.UUID
	Status      string
	Signatures  int
	Goal        int
	EndDate     time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Location    *Location
}
