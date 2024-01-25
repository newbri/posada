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

type ListUsersParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

const getAllUserQuery = `SELECT username, hashed_password, full_name, email, password_changed_at, users.created_at, role_id
FROM users INNER JOIN role on users.role_id = role.internal_id WHERE role.name=$1 LIMIT $2 OFFSET $3;`

func (q *Queries) GetAllCustomer(ctx context.Context, arg ListUsersParams) ([]*User, error) {
	rows, err := q.db.QueryContext(ctx, getAllUserQuery, RoleCustomer, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var items []*User
	for rows.Next() {
		var user User
		var role Role
		if err := rows.Scan(
			&user.Username,
			&user.HashedPassword,
			&user.FullName,
			&user.Email,
			&user.PasswordChangedAt,
			&user.CreatedAt,
			&role.InternalID,
		); err != nil {
			return nil, err
		}
		user.Role, err = q.GetRoleByUUID(ctx, role.InternalID)
		if err != nil {
			return nil, err
		}
		items = append(items, &user)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
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
RETURNING username, hashed_password, full_name, email, password_changed_at, created_at, role_id;
`

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (*User, error) {
	row := q.db.QueryRowContext(ctx, updateUser,
		arg.HashedPassword,
		arg.PasswordChangedAt,
		arg.FullName,
		arg.Email,
		arg.Username,
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

const deleteUserQuery = `DELETE FROM users WHERE username = $1 
     RETURNING username, hashed_password, full_name, email, password_changed_at, created_at, role_id;`

func (q *Queries) DeleteUser(ctx context.Context, username string) (*User, error) {
	row := q.db.QueryRowContext(ctx, deleteUserQuery, username)
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
