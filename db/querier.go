package db

import (
	"context"
	"github.com/google/uuid"
)

type Querier interface {
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	GetUser(ctx context.Context, username string) (User, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
	DeleteUser(ctx context.Context, username string) (User, error)
	GetSession(ctx context.Context, id uuid.UUID) (Session, error)
	CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error)
	CreateRole(ctx context.Context, arg CreateRoleParams) (Role, error)
}
