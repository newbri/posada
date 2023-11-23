package db

import (
	"context"
	"database/sql"
)

const createUser = `
INSERT INTO users (username, hashed_password, full_name, email) 
VALUES ($1,$2,$3,$4)
RETURNING username,hashed_password,full_name,email,password_changed_at,created_at
`

type CreateUserParams struct {
	Username       string `json:"username"`
	HashedPassword string `json:"hashed_password"`
	FullName       string `json:"full_name"`
	Email          string `json:"email"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.Username,
		arg.HashedPassword,
		arg.FullName,
		arg.Email,
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

const getUser = `
SELECT username, hashed_password, full_name, email, password_changed_at, created_at 
FROM users WHERE username = $1;
`

func (q *Queries) GetUser(ctx context.Context, username string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUser, username)
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
