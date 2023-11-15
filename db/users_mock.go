package db

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/mock"
)

type mockUsers struct {
	mock.Mock
	db *sql.DB
}

func (m *mockUsers) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	ret := m.Called(ctx, arg)
	r0 := ret.Get(0).(User)
	return r0, ret.Error(1)
}
