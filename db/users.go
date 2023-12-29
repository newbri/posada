package db

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
)

const createUser = `
INSERT INTO users (username, hashed_password, full_name, email, role_id) 
VALUES ($1,$2,$3,$4,$5)
RETURNING username,hashed_password,full_name,email,password_changed_at,created_at,role_id
`

type CreateUserParams struct {
	Username       string    `json:"username"`
	HashedPassword string    `json:"hashed_password"`
	FullName       string    `json:"full_name"`
	Email          string    `json:"email"`
	RoleID         uuid.UUID `json:"-"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (*User, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.Username,
		arg.HashedPassword,
		arg.FullName,
		arg.Email,
		arg.RoleID,
	)
	var user User
	var role Role
	err := row.Scan(
		&user.Username,
		&user.HashedPassword,
		&user.FullName,
		&user.Email,
		&user.PasswordChangedAt,
		&user.CreatedAt,
		&role.InternalID,
	)
	if err != nil {
		return nil, err
	}
	user.Role, err = q.GetRoleByUUID(ctx, role.InternalID)
	return &user, err
}

const getUser = `
SELECT username, hashed_password, full_name, email, password_changed_at, created_at, role_id 
FROM users WHERE username = $1;
`

func (q *Queries) GetUser(ctx context.Context, username string) (*User, error) {
	row := q.db.QueryRowContext(ctx, getUser, username)
	var user User
	var role Role
	err := row.Scan(
		&user.Username,
		&user.HashedPassword,
		&user.FullName,
		&user.Email,
		&user.PasswordChangedAt,
		&user.CreatedAt,
		&role.InternalID,
	)
	if err != nil {
		return nil, err
	}
	user.Role, err = q.GetRoleByUUID(ctx, role.InternalID)
	return &user, err
}

type UpdateUserParams struct {
	HashedPassword    sql.NullString `json:"hashed_password"`
	PasswordChangedAt sql.NullTime   `json:"password_changed_at"`
	FullName          sql.NullString `json:"full_name"`
	Email             sql.NullString `json:"email"`
	Username          string         `json:"username"`
}

const updateUser = `
UPDATE users
SET hashed_password = coalesce($1, hashed_password),
    password_changed_at = coalesce($2, password_changed_at),
    full_name = coalesce($3, full_name),
    email = coalesce($4, email)
WHERE username = $5
RETURNING username, hashed_password, full_name, email, password_changed_at, created_at;
`

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, updateUser,
		arg.HashedPassword,
		arg.PasswordChangedAt,
		arg.FullName,
		arg.Email,
		arg.Username,
	)
	var user User
	err := row.Scan(
		&user.Username,
		&user.HashedPassword,
		&user.FullName,
		&user.Email,
		&user.PasswordChangedAt,
		&user.CreatedAt,
	)
	return user, err
}

const deleteUserQuery = `DELETE FROM users WHERE username = $1 
     RETURNING username, hashed_password, full_name, email, password_changed_at, created_at;`

func (q *Queries) DeleteUser(ctx context.Context, username string) (User, error) {
	row := q.db.QueryRowContext(ctx, deleteUserQuery, username)
	var user User
	err := row.Scan(
		&user.Username,
		&user.HashedPassword,
		&user.FullName,
		&user.Email,
		&user.PasswordChangedAt,
		&user.CreatedAt,
	)
	return user, err
}
