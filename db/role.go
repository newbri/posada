package db

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"time"
)

const (
	RoleAdmin    = "admin"
	RoleCustomer = "customer"
	RoleRoot     = "root"
)
const createRoleQuery = `
INSERT INTO role (internal_id, name, description, external_id) 
VALUES (gen_random_uuid(),$2,$3, CONCAT('URE',nextval('role_sequence')))
RETURNING internal_id,name,description,external_id,created_at,updated_at
`

type CreateRoleParams struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (q *Queries) CreateRole(ctx context.Context, arg CreateRoleParams) (*Role, error) {
	row := q.db.QueryRowContext(
		ctx,
		createRoleQuery,
		arg.Name,
		arg.Description,
	)
	var role Role
	err := row.Scan(
		&role.InternalID,
		&role.Name,
		&role.Description,
		&role.ExternalID,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	return &role, err
}

type ListRoleParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

const getAllRoleQuery = `
SELECT internal_id,name,description,external_id,created_at,updated_at FROM role LIMIT $1 OFFSET $2;
`

func (q *Queries) GetAllRole(ctx context.Context, arg ListRoleParams) ([]*Role, error) {
	rows, err := q.db.QueryContext(ctx, getAllRoleQuery, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var items []*Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(
			&role.InternalID,
			&role.Name,
			&role.Description,
			&role.ExternalID,
			&role.CreatedAt,
			&role.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, &role)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRoleQuery = `
	SELECT internal_id,name,description,external_id,created_at,updated_at FROM role WHERE external_id = $1;
`

func (q *Queries) GetRole(ctx context.Context, externalId string) (*Role, error) {
	row := q.db.QueryRowContext(ctx, getRoleQuery, externalId)
	var role Role
	err := row.Scan(
		&role.InternalID,
		&role.Name,
		&role.Description,
		&role.ExternalID,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	return &role, err
}

const getRoleByNameQuery = `
	SELECT internal_id,name,description,external_id,created_at,updated_at FROM role WHERE name = $1;`

func (q *Queries) GetRoleByName(ctx context.Context, name string) (*Role, error) {
	row := q.db.QueryRowContext(ctx, getRoleByNameQuery, name)
	var role Role
	err := row.Scan(
		&role.InternalID,
		&role.Name,
		&role.Description,
		&role.ExternalID,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	return &role, err
}

const getRoleByUUIDQuery = `
	SELECT internal_id,name,description,external_id,created_at,updated_at FROM role WHERE internal_id = $1;
`

func (q *Queries) GetRoleByUUID(ctx context.Context, internalId uuid.UUID) (*Role, error) {
	row := q.db.QueryRowContext(ctx, getRoleByUUIDQuery, internalId)
	var role Role
	err := row.Scan(
		&role.InternalID,
		&role.Name,
		&role.Description,
		&role.ExternalID,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	return &role, err
}

type UpdateRoleParams struct {
	ExternalID  string         `json:"external_id"`
	Name        sql.NullString `json:"name"`
	Description sql.NullString `json:"description"`
	UpdateAt    time.Time      `json:"-"`
}

const updateRoleQuery = `
UPDATE role
SET name = coalesce($1, name),
    description = coalesce($2, description),
    updated_at = coalesce($3, updated_at)
WHERE external_id = $4
RETURNING internal_id, name, description, external_id, created_at, updated_at;
`

func (q *Queries) UpdateRole(ctx context.Context, arg UpdateRoleParams) (*Role, error) {
	row := q.db.QueryRowContext(ctx, updateRoleQuery,
		arg.Name,
		arg.Description,
		arg.UpdateAt,
		arg.ExternalID,
	)
	var role Role
	err := row.Scan(
		&role.InternalID,
		&role.Name,
		&role.Description,
		&role.ExternalID,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	return &role, err
}

const deleteRoleQuery = `DELETE FROM role WHERE external_id = $1 
     RETURNING internal_id, name, description, external_id, created_at, updated_at;`

func (q *Queries) DeleteRole(ctx context.Context, externalID string) (*Role, error) {
	row := q.db.QueryRowContext(ctx, deleteRoleQuery, externalID)
	var role Role
	err := row.Scan(
		&role.InternalID,
		&role.Name,
		&role.Description,
		&role.ExternalID,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	return &role, err
}
