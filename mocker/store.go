package mocker

import (
	"context"
	"github.com/newbri/posadamissportia/db"
	"github.com/newbri/posadamissportia/token"
	"github.com/stretchr/testify/mock"
	"time"
)

type TestMocker struct {
	mock.Mock
	db.Querier
	token.Maker
}

func (m *TestMocker) GetRoleByName(ctx context.Context, name string) (*db.Role, error) {
	args := m.Called(ctx, name)
	ret0, _ := args.Get(0).(*db.Role)
	ret1, _ := args.Get(1).(error)
	return ret0, ret1
}

func (m *TestMocker) CreateUser(ctx context.Context, arg *db.CreateUserParams) (*db.User, error) {
	args := m.Called(ctx, arg)
	ret0, _ := args.Get(0).(*db.User)
	ret1, _ := args.Get(1).(error)
	return ret0, ret1
}

func (m *TestMocker) GetUser(ctx context.Context, username string) (*db.User, error) {
	args := m.Called(ctx, username)
	ret0, _ := args.Get(0).(*db.User)
	ret1, _ := args.Get(1).(error)
	return ret0, ret1
}

func (m *TestMocker) DeleteUser(ctx context.Context, username string, deletedAt time.Time) (*db.User, error) {
	args := m.Called(ctx, username, deletedAt)
	ret0, _ := args.Get(0).(*db.User)
	ret1, _ := args.Get(1).(error)
	return ret0, ret1
}

func (m *TestMocker) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (*db.User, error) {
	args := m.Called(ctx, arg)
	ret0, _ := args.Get(0).(*db.User)
	ret1, _ := args.Get(1).(error)
	return ret0, ret1
}

func (m *TestMocker) CreateToken(username string, role *db.Role, duration time.Duration) (string, *token.Payload, error) {
	args := m.Called(username, role, duration)
	ret0, _ := args.Get(0).(string)
	ret1, _ := args.Get(1).(*token.Payload)
	ret2, _ := args.Get(2).(error)
	return ret0, ret1, ret2
}

func (m *TestMocker) VerifyToken(tok string) (*token.Payload, error) {
	args := m.Called(tok)
	ret0, _ := args.Get(0).(*token.Payload)
	ret1, _ := args.Get(1).(error)
	return ret0, ret1
}

func (m *TestMocker) CreateSession(ctx context.Context, arg db.CreateSessionParams) (*db.Session, error) {
	args := m.Called(ctx, arg)
	ret0, _ := args.Get(0).(*db.Session)
	ret1, _ := args.Get(1).(error)
	return ret0, ret1
}

func (m *TestMocker) GetAllCustomer(ctx context.Context, arg db.ListUsersParams) ([]*db.User, error) {
	args := m.Called(ctx, arg)
	ret0, _ := args.Get(0).([]*db.User)
	ret1, _ := args.Get(1).(error)
	return ret0, ret1
}

func (m *TestMocker) GetAllAdmin(ctx context.Context, arg db.ListUsersParams) ([]*db.User, error) {
	args := m.Called(ctx, arg)
	ret0, _ := args.Get(0).([]*db.User)
	ret1, _ := args.Get(1).(error)
	return ret0, ret1
}

func (m *TestMocker) CreateRole(ctx context.Context, arg db.CreateRoleParams) (*db.Role, error) {
	args := m.Called(ctx, arg)
	ret0, _ := args.Get(0).(*db.Role)
	ret1, _ := args.Get(1).(error)
	return ret0, ret1
}

func (m *TestMocker) GetAllRole(ctx context.Context, arg db.ListRoleParams) ([]*db.Role, error) {
	args := m.Called(ctx, arg)
	ret0, _ := args.Get(0).([]*db.Role)
	ret1, _ := args.Get(1).(error)
	return ret0, ret1
}

func (m *TestMocker) GetRole(ctx context.Context, externalId string) (*db.Role, error) {
	args := m.Called(ctx, externalId)
	ret0, _ := args.Get(0).(*db.Role)
	ret1, _ := args.Get(1).(error)
	return ret0, ret1
}

func (m *TestMocker) UpdateRole(ctx context.Context, arg db.UpdateRoleParams) (*db.Role, error) {
	args := m.Called(ctx, arg)
	ret0, _ := args.Get(0).(*db.Role)
	ret1, _ := args.Get(1).(error)
	return ret0, ret1
}

func (m *TestMocker) DeleteRole(ctx context.Context, externalID string) (*db.Role, error) {
	args := m.Called(ctx, externalID)
	ret0, _ := args.Get(0).(*db.Role)
	ret1, _ := args.Get(1).(error)
	return ret0, ret1
}
