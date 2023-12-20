package db

import (
	"context"
	"github.com/google/uuid"
)

const createRole = `
INSERT INTO role (id, name, description) 
VALUES ($1,$2,$3)
RETURNING id,name,description,created_at,updated_at
`

type CreateRoleParams struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

func (q *Queries) CreateRole(ctx context.Context, arg CreateRoleParams) (Role, error) {
	row := q.db.QueryRowContext(
		ctx,
		createRole,
		arg.ID,
		arg.Name,
		arg.Description,
	)
	var role Role
	err := row.Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	return role, err
}

type ListRoleParams struct {
	Limit  int32 `json:"limit""`
	Offset int32 `json:"offset"`
}

const getAllRole = `
SELECT id,name,description,created_at,updated_at FROM role LIMIT $1 OFFSET $2;
`

func (q *Queries) GetAllRole(ctx context.Context, arg ListRoleParams) ([]*Role, error) {
	rows, err := q.db.QueryContext(ctx, getAllRole, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
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
