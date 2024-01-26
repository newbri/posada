package db

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	Username          string    `json:"username"`
	HashedPassword    string    `json:"-"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
	Role              *Role     `json:"role"`
}

type Session struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	ClientIp     string    `json:"client_ip"`
	IsBlocked    bool      `json:"is_blocked"`
	ExpiredAt    time.Time `json:"expired_at"`
	CreatedAt    time.Time `json:"created_at"`
	BlockedAt    time.Time `json:"blocked_at"`
}

type Role struct {
	InternalID  uuid.UUID `json:"-"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ExternalID  string    `json:"external_id"`
	UpdatedAt   time.Time `json:"expired_at"`
	CreatedAt   time.Time `json:"created_at"`
}
