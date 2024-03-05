package db

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"time"
)

const insertUserQuery = `
INSERT INTO users (username, hashed_password, full_name, email, role_id) 
VALUES ($1,$2,$3,$4,$5)
RETURNING username, hashed_password, full_name, email, password_changed_at, created_at, role_id, is_deleted, deleted_at;
`

type CreateUserParams struct {
	Username       string    `json:"username"`
	HashedPassword string    `json:"hashed_password"`
	FullName       string    `json:"full_name"`
	Email          string    `json:"email"`
	RoleID         uuid.UUID `json:"-"`
}

func (q *Queries) CreateUser(ctx context.Context, arg *CreateUserParams) (*User, error) {
	row := q.db.QueryRowContext(ctx, insertUserQuery,
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
		&user.IsDeleted,
		&user.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	user.Role, err = q.GetRoleByUUID(ctx, role.InternalID)
	return &user, err
}

const getUserQuery = `
SELECT username, hashed_password, full_name, email, password_changed_at,
       users.created_at, is_deleted, deleted_at,
       p.internal_id, p.name, p.description, p.external_id, p.created_at, p.updated_at
FROM users FULL OUTER JOIN role p ON users.role_id = p.internal_id 
WHERE username = $1 AND is_deleted = $2;
`

func (q *Queries) GetUser(ctx context.Context, username string) (*User, error) {
	row := q.db.QueryRowContext(ctx, getUserQuery, username, false)
	var user User
	var role Role
	err := row.Scan(
		&user.Username,
		&user.HashedPassword,
		&user.FullName,
		&user.Email,
		&user.PasswordChangedAt,
		&user.CreatedAt,
		&user.IsDeleted,
		&user.DeletedAt,
		&role.InternalID,
		&role.Name,
		&role.Description,
		&role.ExternalID,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	user.Role = &role
	//user.Role, err = q.GetRoleByUUID(ctx, role.InternalID)
	return &user, err
}

type ListUsersParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

const getAllUserQuery = `SELECT username, hashed_password, full_name, email, password_changed_at, users.created_at, role_id, is_deleted, deleted_at
FROM users INNER JOIN role on users.role_id = role.internal_id WHERE role.name = $1 AND is_deleted = $2 LIMIT $3 OFFSET $4;`

func (q *Queries) GetAllCustomer(ctx context.Context, arg ListUsersParams) ([]*User, error) {
	return q.getUsersByRole(ctx, getAllUserQuery, RoleCustomer, false, arg)
}

func (q *Queries) GetAllAdmin(ctx context.Context, arg ListUsersParams) ([]*User, error) {
	return q.getUsersByRole(ctx, getAllUserQuery, RoleAdmin, false, arg)
}

type UpdateUserParams struct {
	HashedPassword    sql.NullString `json:"hashed_password"`
	PasswordChangedAt sql.NullTime   `json:"password_changed_at"`
	FullName          sql.NullString `json:"full_name"`
	Email             sql.NullString `json:"email"`
	Username          string         `json:"username"`
}

const updateUserQuery = `
UPDATE users
SET hashed_password = coalesce($1, hashed_password),
    password_changed_at = coalesce($2, password_changed_at),
    full_name = coalesce($3, full_name),
    email = coalesce($4, email)
WHERE username = $5 AND is_deleted = $6
RETURNING username, hashed_password, full_name, email, password_changed_at, created_at, role_id, is_deleted, deleted_at;
`

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (*User, error) {
	row := q.db.QueryRowContext(ctx, updateUserQuery,
		arg.HashedPassword,
		arg.PasswordChangedAt,
		arg.FullName,
		arg.Email,
		arg.Username,
		false,
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
		&user.IsDeleted,
		&user.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	user.Role, err = q.GetRoleByUUID(ctx, role.InternalID)
	return &user, err
}

const deleteUserQuery = `UPDATE users SET is_deleted = $1, deleted_at = $2 WHERE username = $3 AND is_deleted = $4
     RETURNING username, hashed_password, full_name, email, password_changed_at, created_at, role_id, is_deleted, deleted_at;`

func (q *Queries) DeleteUser(ctx context.Context, username string, deletedAt time.Time) (*User, error) {
	row := q.db.QueryRowContext(ctx, deleteUserQuery, true, deletedAt, username, false)
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
		&user.IsDeleted,
		&user.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	user.Role, err = q.GetRoleByUUID(ctx, role.InternalID)
	return &user, err
}

func (q *Queries) getUsersByRole(ctx context.Context, query string, role string, isDeleted bool, arg ListUsersParams) ([]*User, error) {
	rows, err := q.db.QueryContext(ctx, query, role, isDeleted, arg.Limit, arg.Offset)
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
			&user.IsDeleted,
			&user.DeletedAt,
		); err != nil {
			return nil, err
		}
		user.Role, err = q.GetRoleByUUID(ctx, role.InternalID)
		if err != nil {
			return nil, err
		}
		items = append(items, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
