package models

import (
	"time"

	"github.com/google/uuid"
)

type Proposal struct {
	ID           uuid.UUID
	CityID       uuid.UUID
	Title        string
	Description  string
	Status       string
	InitiatorID  uuid.UUID
	AddressToID  *uuid.UUID
	AgreedNum    int
	DisagreedNum int
	EndDate      time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Location     *Location
}
