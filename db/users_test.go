package db

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestCreateUser(t *testing.T) {
	store := NewStore(db)
	user := createRandomUser(t)

	rows := sqlmock.NewRows([]string{"username", "hashed_password", "full_name", "email", "password_changed_at", "created_at"}).
		AddRow(user.Username,
			user.HashedPassword,
			user.FullName,
			user.Email,
			user.PasswordChangedAt,
			user.CreatedAt,
		)

	testCases := []struct {
		name     string
		query    string
		arg      CreateUserParams
		validate func(query string, arg CreateUserParams)
	}{
		{
			name:  "OK",
			query: createUser,
			arg: CreateUserParams{
				Username:       user.Username,
				HashedPassword: user.HashedPassword,
				FullName:       user.FullName,
				Email:          user.Email,
			},
			validate: func(query string, arg CreateUserParams) {

				mocker.ExpectQuery(regexp.QuoteMeta(createUser)).
					WithArgs(arg.Username, arg.HashedPassword, arg.FullName, arg.Email).
					WillReturnRows(rows)

				actualUser, err := store.CreateUser(context.Background(), arg)
				require.NoError(t, err)
				require.NotNil(t, actualUser)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			tc.validate(tc.query, tc.arg)

			// we make sure that all expectations were met
			if err := mocker.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	store := NewStore(db)
	user := createRandomUser(t)

	rows := sqlmock.NewRows([]string{"username", "hashed_password", "full_name", "email", "password_changed_at", "created_at"}).
		AddRow(user.Username,
			user.HashedPassword,
			user.FullName,
			user.Email,
			user.PasswordChangedAt,
			user.CreatedAt,
		)

	testCases := []struct {
		name     string
		query    string
		validate func(query, username string)
	}{
		{
			name:  "OK",
			query: getUser,
			validate: func(query, username string) {

				mocker.ExpectQuery(regexp.QuoteMeta(getUser)).
					WithArgs(username).
					WillReturnRows(rows)

				actualUser, err := store.GetUser(context.Background(), username)
				require.NoError(t, err)
				require.NotNil(t, actualUser)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			tc.validate(tc.query, user.Username)

			// we make sure that all expectations were met
			if err := mocker.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestUpdateUserFullName(t *testing.T) {
	store := NewStore(db)
	user := createRandomUser(t)

	arg := UpdateUserParams{
		Username: user.Username,
		FullName: sql.NullString{
			String: user.FullName,
			Valid:  true,
		},
	}

	mockUpdateUserDB(user, arg)
	updatedUser, err := store.UpdateUser(context.Background(), arg)

	require.NoError(t, err)
	require.Equal(t, user.Email, updatedUser.Email)
	require.Equal(t, user.FullName, updatedUser.FullName)
	require.Equal(t, user.HashedPassword, updatedUser.HashedPassword)
}

func TestUpdateUserEmail(t *testing.T) {
	store := NewStore(db)
	user := createRandomUser(t)

	arg := UpdateUserParams{
		Username: user.Username,
		Email: sql.NullString{
			String: user.Email,
			Valid:  true,
		},
	}

	mockUpdateUserDB(user, arg)
	updatedUser, err := store.UpdateUser(context.Background(), arg)

	require.NoError(t, err)
	require.Equal(t, user.Email, updatedUser.Email)
	require.Equal(t, user.FullName, updatedUser.FullName)
	require.Equal(t, user.HashedPassword, updatedUser.HashedPassword)
}

func mockUpdateUserDB(user User, args UpdateUserParams) {
	rows := sqlmock.NewRows([]string{"username", "hashed_password", "full_name", "email", "password_changed_at", "created_at"}).
		AddRow(user.Username,
			user.HashedPassword,
			user.FullName,
			user.Email,
			user.PasswordChangedAt,
			user.CreatedAt,
		)

	mocker.ExpectQuery(regexp.QuoteMeta(updateUser)).
		WithArgs(args.HashedPassword, args.PasswordChangedAt, args.FullName, args.Email, args.Username).
		WillReturnRows(rows)
}
