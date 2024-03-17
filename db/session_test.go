package db

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestCreateSession(t *testing.T) {
	db, mocker, err := sqlmock.New()
	require.NoError(t, err)
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	arg := createSessionParams()
	expectedSession := createSession(arg)
	testCases := []struct {
		name             string
		sessionQueryRows *sqlmock.Rows
		mock             func(userQueryRows *sqlmock.Rows, arg *CreateSessionParams)
		response         func(querier Querier, expectedSession *Session, arg CreateSessionParams)
	}{
		{
			name:             "OK",
			sessionQueryRows: getMockedExpectedCreateSession(expectedSession),
			mock: func(userQueryRows *sqlmock.Rows, arg *CreateSessionParams) {
				// the CreateUser sql mock
				mocker.
					ExpectQuery(regexp.QuoteMeta(createSessionQuery)).
					WithArgs(
						arg.ID,
						arg.Username,
						arg.RefreshToken,
						arg.UserAgent,
						arg.ClientIp,
						arg.IsBlocked,
						arg.ExpiredAt,
						arg.CreatedAt,
					).
					WillReturnRows(userQueryRows)
			},
			response: func(querier Querier, expectedSession *Session, arg CreateSessionParams) {
				actualSession, err := querier.CreateSession(context.Background(), arg)
				require.NoError(t, err)
				require.Equal(t, actualSession, expectedSession)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockQuery := &Queries{db: db}

			tc.mock(tc.sessionQueryRows, arg)
			tc.response(mockQuery, expectedSession, *arg)
		})
	}
}
