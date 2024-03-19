package db

import (
	"github.com/google/uuid"
	"time"
)

type CreatePropertyParams struct {
	InternalID uuid.UUID
	ExternalID string
	Name       string
	Address    string
	State      string
	Country    string
	PostalCode string
	Phone      string
	Email      string
	ExpiredAt  time.Time
	CreatedAt  time.Time
}
