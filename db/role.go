package db

import (
	"context"
	"github.com/google/uuid"
)

const createRole = `
INSERT INTO role (id, name, description) 
VALUES ($1,$2,$3)
RETURNING id,name,description,created_at,updated_at;
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
