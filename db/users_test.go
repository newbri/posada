package db

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/newbri/posadamissportia/db/util"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
	"time"
)

type mockQuerierDB struct {
	mock.Mock
	*Queries
}

func (m *mockQuerierDB) CreateUser(ctx context.Context, arg *CreateUserParams) (*User, error) {
	return m.Queries.CreateUser(ctx, arg)
}

func (m *mockQuerierDB) GetRoleByUUID(ctx context.Context, internalId uuid.UUID) (*Role, error) {
	return m.Queries.GetRoleByUUID(ctx, internalId)
}

func (m *mockQuerierDB) GetUser(ctx context.Context, username string) (*User, error) {
	return m.Queries.GetUser(ctx, username)
}

func (m *mockQuerierDB) UpdateUser(ctx context.Context, arg UpdateUserParams) (*User, error) {
	return m.Queries.UpdateUser(ctx, arg)
}

func (m *mockQuerierDB) DeleteUser(ctx context.Context, username string, deletedAt time.Time) (*User, error) {
	return m.Queries.DeleteUser(ctx, username, deletedAt)
}

func TestCreateUser(t *testing.T) {
	expectedUser := createRandomUserWithRole(RoleCustomer, false)
	testCases := []struct {
		name          string
		userQueryRows *sqlmock.Rows
		roleQueryRows *sqlmock.Rows
		mock          func(userQueryRows *sqlmock.Rows, roleQueryRows *sqlmock.Rows, arg *CreateUserParams, roleId uuid.UUID)
		response      func(querier Querier, expectedUser *User, arg *CreateUserParams)
	}{
		{
			name:          "OK",
			userQueryRows: getMockedExpectedUserRows(expectedUser),
			roleQueryRows: getMockedExpectedRoleRows(expectedUser.Role),
			mock: func(userQueryRows *sqlmock.Rows, roleQueryRows *sqlmock.Rows, arg *CreateUserParams, roleId uuid.UUID) {
				// the CreateUser sql mock
				mocker.
					ExpectQuery(regexp.QuoteMeta(insertUserQuery)).
					WithArgs(
						arg.Username,
						arg.HashedPassword,
						arg.FullName,
						arg.Email,
						arg.RoleID,
					).
					WillReturnRows(userQueryRows)

				// the GetRoleByUUID sql mock
				mocker.ExpectQuery(regexp.QuoteMeta(getRoleByUUIDQuery)).
					WithArgs(roleId).
					WillReturnRows(roleQueryRows)

			},
			response: func(querier Querier, expectedUser *User, arg *CreateUserParams) {
				actualUser, err := querier.CreateUser(context.Background(), arg)
				require.NoError(t, err)
				require.Equal(t, actualUser, expectedUser)
				require.Equal(t, actualUser.Role, expectedUser.Role)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockQuery := &mockQuerierDB{Queries: New(db)}

			arg := createUserParams(expectedUser)
			tc.mock(tc.userQueryRows, tc.roleQueryRows, arg, expectedUser.Role.InternalID)
			tc.response(mockQuery, expectedUser, arg)
		})
	}
}

func TestGetUser(t *testing.T) {
	expectedUser := createRandomUserWithRole(RoleCustomer, false)
	testCases := []struct {
		name          string
		userQueryRows *sqlmock.Rows
		roleQueryRows *sqlmock.Rows
		mock          func(userQueryRows *sqlmock.Rows, roleQueryRows *sqlmock.Rows, username string, roleId uuid.UUID, isDeleted bool)
		response      func(querier Querier, expectedUser *User)
	}{
		{
			name:          "OK",
			userQueryRows: getMockedExpectedUserRows(expectedUser),
			roleQueryRows: getMockedExpectedRoleRows(expectedUser.Role),
			mock: func(userQueryRows *sqlmock.Rows, roleQueryRows *sqlmock.Rows, username string, roleId uuid.UUID, isDeleted bool) {

				// the CreateUser sql mock
				mocker.
					ExpectQuery(regexp.QuoteMeta(getUserQuery)).
					WithArgs(username, isDeleted).
					WillReturnRows(userQueryRows)

				// the GetRoleByUUID sql mock
				mocker.ExpectQuery(regexp.QuoteMeta(getRoleByUUIDQuery)).
					WithArgs(roleId).
					WillReturnRows(roleQueryRows)
			},
			response: func(querier Querier, expectedUser *User) {
				actualUser, err := querier.GetUser(context.Background(), expectedUser.Username)
				require.NoError(t, err)
				require.Equal(t, actualUser, expectedUser)
				require.Equal(t, actualUser.Role, expectedUser.Role)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockQuery := &mockQuerierDB{Queries: New(db)}

			tc.mock(tc.userQueryRows, tc.roleQueryRows, expectedUser.Username, expectedUser.Role.InternalID, false)
			tc.response(mockQuery, expectedUser)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	expectedUser := createRandomUserWithRole(RoleCustomer, false)
	testCases := []struct {
		name          string
		userQueryRows *sqlmock.Rows
		roleQueryRows *sqlmock.Rows
		arg           UpdateUserParams
		mock          func(userQueryRows *sqlmock.Rows, roleQueryRows *sqlmock.Rows, arg UpdateUserParams, roleId uuid.UUID, isDeleted bool)
		response      func(querier Querier, arg UpdateUserParams)
	}{
		{
			name:          "OK FullName",
			userQueryRows: getMockedExpectedUserRows(expectedUser),
			roleQueryRows: getMockedExpectedRoleRows(expectedUser.Role),
			arg: UpdateUserParams{
				Username: expectedUser.Username,
				FullName: sql.NullString{
					String: util.RandomString(8),
					Valid:  true,
				},
			},
			mock: func(userQueryRows *sqlmock.Rows, roleQueryRows *sqlmock.Rows, arg UpdateUserParams, roleId uuid.UUID, isDeleted bool) {
				// the CreateUser sql mock
				mocker.
					ExpectQuery(regexp.QuoteMeta(updateUserQuery)).
					WithArgs(arg.HashedPassword, arg.PasswordChangedAt, arg.FullName, arg.Email, arg.Username, isDeleted).
					WillReturnRows(userQueryRows)

				// the GetRoleByUUID sql mock
				mocker.ExpectQuery(regexp.QuoteMeta(getRoleByUUIDQuery)).
					WithArgs(roleId).
					WillReturnRows(roleQueryRows)
			},
			response: func(querier Querier, arg UpdateUserParams) {
				actualUser, err := querier.UpdateUser(context.Background(), arg)
				require.NoError(t, err)
				require.Equal(t, actualUser, expectedUser)
				require.Equal(t, actualUser.Role, expectedUser.Role)
			},
		},
		{
			name:          "OK Email",
			userQueryRows: getMockedExpectedUserRows(expectedUser),
			roleQueryRows: getMockedExpectedRoleRows(expectedUser.Role),
			arg: UpdateUserParams{
				Username: expectedUser.Username,
				Email: sql.NullString{
					String: util.RandomEmail(),
					Valid:  true,
				},
			},
			mock: func(userQueryRows *sqlmock.Rows, roleQueryRows *sqlmock.Rows, arg UpdateUserParams, roleId uuid.UUID, isDeleted bool) {
				// the CreateUser sql mock
				mocker.
					ExpectQuery(regexp.QuoteMeta(updateUserQuery)).
					WithArgs(arg.HashedPassword, arg.PasswordChangedAt, arg.FullName, arg.Email, arg.Username, isDeleted).
					WillReturnRows(userQueryRows)

				// the GetRoleByUUID sql mock
				mocker.ExpectQuery(regexp.QuoteMeta(getRoleByUUIDQuery)).
					WithArgs(roleId).
					WillReturnRows(roleQueryRows)
			},
			response: func(querier Querier, arg UpdateUserParams) {
				actualUser, err := querier.UpdateUser(context.Background(), arg)
				require.NoError(t, err)
				require.Equal(t, actualUser, expectedUser)
				require.Equal(t, actualUser.Role, expectedUser.Role)
			},
		},
		{
			name:          "Error",
			userQueryRows: getMockedWrongExpectedUserRows(expectedUser),
			roleQueryRows: getMockedExpectedRoleRows(expectedUser.Role),
			arg: UpdateUserParams{
				Username: expectedUser.Username,
				FullName: sql.NullString{
					String: util.RandomString(8),
					Valid:  true,
				},
			},
			mock: func(userQueryRows *sqlmock.Rows, roleQueryRows *sqlmock.Rows, arg UpdateUserParams, roleId uuid.UUID, isDeleted bool) {
				// the CreateUser sql mock
				mocker.
					ExpectQuery(regexp.QuoteMeta(updateUserQuery)).
					WithArgs(arg.HashedPassword, arg.PasswordChangedAt, arg.FullName, arg.Email, arg.Username, isDeleted).
					WillReturnRows(userQueryRows)
			},
			response: func(querier Querier, arg UpdateUserParams) {
				actualUser, err := querier.UpdateUser(context.Background(), arg)
				require.Error(t, err)
				require.Nil(t, actualUser)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockQuery := &mockQuerierDB{Queries: New(db)}

			tc.mock(tc.userQueryRows, tc.roleQueryRows, tc.arg, expectedUser.Role.InternalID, false)
			tc.response(mockQuery, tc.arg)
		})
	}
}

func TestDeleteUser(t *testing.T) {
	expectedUser := createRandomUserWithRole(RoleCustomer, false)
	testCases := []struct {
		name          string
		userQueryRows *sqlmock.Rows
		roleQueryRows *sqlmock.Rows
		mock          func(userQueryRows *sqlmock.Rows, roleQueryRows *sqlmock.Rows, username string, t time.Time, roleId uuid.UUID, isDeleted bool)
		response      func(querier Querier, deletedAt time.Time)
	}{
		{
			name:          "OK FullName",
			userQueryRows: getMockedExpectedUserRows(expectedUser),
			roleQueryRows: getMockedExpectedRoleRows(expectedUser.Role),
			mock: func(userQueryRows *sqlmock.Rows, roleQueryRows *sqlmock.Rows, username string, t time.Time, roleId uuid.UUID, isDeleted bool) {
				// the CreateUser sql mock
				mocker.
					ExpectQuery(regexp.QuoteMeta(deleteUserQuery)).
					WithArgs(true, t, username, false).
					WillReturnRows(userQueryRows)

				// the GetRoleByUUID sql mock
				mocker.ExpectQuery(regexp.QuoteMeta(getRoleByUUIDQuery)).
					WithArgs(roleId).
					WillReturnRows(roleQueryRows)
			},
			response: func(querier Querier, deletedAt time.Time) {
				actualUser, err := querier.DeleteUser(context.Background(), expectedUser.Username, deletedAt)
				require.NoError(t, err)
				require.Equal(t, actualUser, expectedUser)
			},
		},
		{
			name:          "Error",
			userQueryRows: getMockedWrongExpectedUserRows(expectedUser),
			roleQueryRows: getMockedExpectedRoleRows(expectedUser.Role),
			mock: func(userQueryRows *sqlmock.Rows, roleQueryRows *sqlmock.Rows, username string, t time.Time, roleId uuid.UUID, isDeleted bool) {
				// the CreateUser sql mock
				mocker.
					ExpectQuery(regexp.QuoteMeta(deleteUserQuery)).
					WithArgs(true, t, username, false).
					WillReturnRows(userQueryRows)
			},
			response: func(querier Querier, deletedAt time.Time) {
				actualUser, err := querier.DeleteUser(context.Background(), expectedUser.Username, deletedAt)
				require.Error(t, err)
				require.Nil(t, actualUser)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockQuery := &mockQuerierDB{Queries: New(db)}

			deletedAt := time.Now()
			tc.mock(tc.userQueryRows, tc.roleQueryRows, expectedUser.Username, deletedAt, expectedUser.Role.InternalID, false)
			tc.response(mockQuery, deletedAt)
		})
	}
}
