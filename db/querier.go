package db

import "context"

type Querier interface {
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
}
