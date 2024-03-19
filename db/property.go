package db

import (
	"context"
	"github.com/google/uuid"
	"time"
)

const createPropertyQuery = `
	INSERT INTO property(internal_id, external_id, name, address, state, country, postal_code, phone, email, expired_at, created_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	RETURNING internal_id, external_id, name, address, state, country, postal_code, phone, email, expired_at, created_at;
`

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

func (q *Queries) CreateProperty(ctx context.Context, arg CreatePropertyParams) (*Property, error) {
	row := q.db.QueryRowContext(
		ctx,
		createPropertyQuery,
		arg.InternalID,
		arg.ExternalID,
		arg.Name,
		arg.Address,
		arg.State,
		arg.Country,
		arg.PostalCode,
		arg.Phone,
		arg.Email,
		arg.ExpiredAt,
		arg.CreatedAt,
	)
	var property Property
	err := row.Scan(
		&property.InternalID,
		&property.ExternalID,
		&property.Name,
		&property.Address,
		&property.State,
		&property.Country,
		&property.PostalCode,
		&property.Phone,
		&property.Email,
		&property.ExpiredAt,
		&property.CreatedAt,
	)
	return &property, err
}
