package db

import (
	"context"
	"github.com/google/uuid"
	"time"
)

const createSessionQuery = `-- name: CreateSession :one
INSERT INTO sessions (id,
                      username,
                      refresh_token,
                      user_agent,
                      client_ip,
                      is_blocked,
                      expired_at,
                      created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, username, refresh_token, user_agent, client_ip, is_blocked, expired_at, created_at
`

type CreateSessionParams struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	ClientIp     string    `json:"client_ip"`
	IsBlocked    bool      `json:"is_blocked"`
	ExpiredAt    time.Time `json:"expired_at"`
	CreatedAt    time.Time `json:"created_at"`
}

func (q *Queries) CreateSession(ctx context.Context, arg CreateSessionParams) (*Session, error) {
	row := q.db.QueryRowContext(ctx, createSessionQuery,
		arg.ID,
		arg.Username,
		arg.RefreshToken,
		arg.UserAgent,
		arg.ClientIp,
		arg.IsBlocked,
		arg.ExpiredAt,
		arg.CreatedAt,
	)
	var session Session
	err := row.Scan(
		&session.ID,
		&session.Username,
		&session.RefreshToken,
		&session.UserAgent,
		&session.ClientIp,
		&session.IsBlocked,
		&session.ExpiredAt,
		&session.CreatedAt,
	)
	return &session, err
}

const getSessionQuery = `-- name: GetSession :one
SELECT id, username, refresh_token, user_agent, client_ip, is_blocked, expired_at, created_at
FROM sessions
WHERE id = $1
LIMIT 1
`

func (q *Queries) GetSession(ctx context.Context, id uuid.UUID) (*Session, error) {
	row := q.db.QueryRowContext(ctx, getSessionQuery, id)
	var session Session
	err := row.Scan(
		&session.ID,
		&session.Username,
		&session.RefreshToken,
		&session.UserAgent,
		&session.ClientIp,
		&session.IsBlocked,
		&session.ExpiredAt,
		&session.CreatedAt,
	)
	return &session, err
}
