package db

import (
	"context"
	"github.com/google/uuid"
	"time"
)

type Querier interface {
	CreateUser(ctx context.Context, arg *CreateUserParams) (*User, error)
	GetUser(ctx context.Context, username string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (*User, error)
	DeleteUser(ctx context.Context, username string, deletedAt time.Time) (*User, error)
	GetSession(ctx context.Context, id uuid.UUID) (*Session, error)
	CreateSession(ctx context.Context, arg CreateSessionParams) (*Session, error)
	BlockSession(ctx context.Context, sessionID uuid.UUID) (*Session, error)
	CreateRole(ctx context.Context, arg CreateRoleParams) (*Role, error)
	GetAllRole(ctx context.Context, arg ListRoleParams) ([]*Role, error)
	GetRole(ctx context.Context, externalId string) (*Role, error)
	GetRoleByName(ctx context.Context, name string) (*Role, error)
	GetRoleByUUID(ctx context.Context, internalId uuid.UUID) (*Role, error)
	UpdateRole(ctx context.Context, arg UpdateRoleParams) (*Role, error)
	DeleteRole(ctx context.Context, externalID string) (*Role, error)
	GetAllCustomer(ctx context.Context, arg ListUsersParams) ([]*User, error)
	GetAllAdmin(ctx context.Context, arg ListUsersParams) ([]*User, error)
	CreateProperty(ctx context.Context, arg CreatePropertyParams) (*Property, error)
	ActivateDeactivateProperty(ctx context.Context, isActive bool, externalId string) (*Property, error)
	GetAllProperty(ctx context.Context, arg LimitOffset) ([]*Property, error)
	GetProperty(ctx context.Context, Id string) (*Property, error)
	UpdateProperty(ctx context.Context, arg UpdatePropertyParams) (*Property, error)
	DeleteProperty(ctx context.Context, externalID string) (*Property, error)
}
