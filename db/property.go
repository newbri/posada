package db

import (
	"context"
	"time"
)

const createPropertyQuery = `
	INSERT INTO property(internal_id, external_id, name, address, state, city, country, postal_code, phone, email, expired_at, created_at) 
	VALUES (gen_random_uuid(), CONCAT('PRO',nextval('property_sequence')), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING internal_id, external_id, name, address, state, city, country, postal_code, phone, email, is_active, expired_at, created_at;
`

type CreatePropertyParams struct {
	Name       string
	Address    string
	City       string
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
		arg.Name,
		arg.Address,
		arg.State,
		arg.City,
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
		&property.City,
		&property.Country,
		&property.PostalCode,
		&property.Phone,
		&property.Email,
		&property.IsActive,
		&property.ExpiredAt,
		&property.CreatedAt,
	)
	return &property, err
}
