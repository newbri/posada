package db

import (
	"context"
	"database/sql"
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
	return getProperty(row)
}

const activatePropertyQuery = `
	UPDATE property SET is_active = $1 WHERE property.external_id = $2
	RETURNING internal_id, external_id, name, address, state, city, country, postal_code, phone, email, is_active, expired_at, created_at;
`

func (q *Queries) ActivateDeactivateProperty(ctx context.Context, isActive bool, externalId string) (*Property, error) {
	row := q.db.QueryRowContext(ctx, activatePropertyQuery, isActive, externalId)
	return getProperty(row)
}

const getAllPropertyQuery = `
SELECT internal_id, external_id, name, address, state, city, country, postal_code, phone, email, is_active, expired_at, created_at
FROM property LIMIT $1 OFFSET $2;
`

func (q *Queries) GetAllProperty(ctx context.Context, arg LimitOffset) ([]*Property, error) {
	rows, err := q.db.QueryContext(ctx, getAllPropertyQuery, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var items []*Property
	for rows.Next() {
		var property Property
		err := rows.Scan(
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
		if err != nil {
			return nil, err
		}
		items = append(items, &property)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPropertyQuery = `
SELECT internal_id, external_id, name, address, state, city, country, postal_code, phone, email, is_active, expired_at, created_at
FROM property WHERE external_id = $1;
`

func (q *Queries) GetProperty(ctx context.Context, Id string) (*Property, error) {
	row := q.db.QueryRowContext(ctx, getPropertyQuery, Id)
	return getProperty(row)
}

type UpdatePropertyParams struct {
	ExternalID string         `json:"external_id"`
	Name       sql.NullString `json:"name"`
	Address    sql.NullString `json:"address"`
	State      sql.NullString `json:"state"`
	City       sql.NullString `json:"city"`
	Country    sql.NullString `json:"country"`
	PostalCode sql.NullString `json:"postal_code"`
	Phone      sql.NullString `json:"phone"`
	Email      sql.NullString `json:"email"`
}

const updatePropertyQuery = `
UPDATE property
SET name = coalesce($1, name), 
    address = coalesce($2, address), 
    state = coalesce($3, state), 
    city = coalesce($4, city), 
    country = coalesce($5, country), 
    postal_code = coalesce($6, postal_code), 
    phone = coalesce($7, phone), 
    email = coalesce($8, email)
WHERE external_id = $9
RETURNING internal_id, external_id, name, address, state, city, country, postal_code, phone, email, is_active, expired_at, created_at;
`

func (q *Queries) UpdateProperty(ctx context.Context, arg UpdatePropertyParams) (*Property, error) {
	row := q.db.QueryRowContext(ctx, updatePropertyQuery,
		arg.Name,
		arg.Address,
		arg.State,
		arg.City,
		arg.Country,
		arg.PostalCode,
		arg.Phone,
		arg.Email,
		arg.ExternalID,
	)
	return getProperty(row)
}

const deletePropertyQuery = `
	DELETE FROM property WHERE external_id = $1
	RETURNING internal_id, external_id, name, address, state, city, country, postal_code, phone, email, is_active, expired_at, created_at;
`

func (q *Queries) DeleteProperty(ctx context.Context, externalID string) (*Property, error) {
	row := q.db.QueryRowContext(ctx, deletePropertyQuery, externalID)
	return getProperty(row)
}

func getProperty(row *sql.Row) (*Property, error) {
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
	if err != nil {
		return nil, err
	}
	return &property, nil
}
