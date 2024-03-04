package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pashagolub/pgxmock/v3"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestGetAllRole tests the GetAllRole function
func TestGetAllRole(t *testing.T) {
	//db, mockDB, err := sqlmock.New()
	//require.NoError(t, err)

	updateDate := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)

	//defer func(db *sql.DB) {
	//	_ = db.Close()
	//}(db)

	// create an instance of our Queries struct.
	//q := New(db)

	// specify the test cases.
	tests := []struct {
		name        string
		mock        func()
		input       ListRoleParams
		expectedErr bool
	}{
		{
			name: "OK",
			mock: func() {
				role := createRole()
				rows := getRow(role, updateDate)
				setupMockBD(db, getAllRoleQuery, rows, nil, 2, 0)
			},
			input:       ListRoleParams{Limit: 2, Offset: 0},
			expectedErr: false,
		},
		{
			name: "Fail",
			mock: func() {
				setupMockBD(mockDB, getAllRoleQuery, nil, sql.ErrConnDone, 2, 0)
			},
			input:       ListRoleParams{Limit: 2, Offset: 0},
			expectedErr: true,
		},
		{
			name: "Close",
			mock: func() {
				mockDB.ExpectClose()
				err = db.Close()
				require.NoError(t, err) // There should be no
				setupMockBD(mockDB, getAllRoleQuery, nil, sql.ErrConnDone, 2, 0)
			},
			input:       ListRoleParams{Limit: 2, Offset: 0},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			_, err := q.GetAllRole(context.Background(), tt.input)

			if tt.expectedErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			err = mockDB.ExpectationsWereMet()
			require.NoError(t, err)
		})
	}
}

func TestQueries_GetRole(t *testing.T) {
	role := createRole()
	updateDate := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		mockFunc func(mock sqlmock.Sqlmock)
		id       string
		wantErr  bool
	}{
		{
			name: "OK",
			mockFunc: func(mock sqlmock.Sqlmock) {
				rows := getRow(role, updateDate)
				setupMockBD(mock, getRoleQuery, rows, nil, role.ExternalID)
			},
			id:      role.ExternalID,
			wantErr: false,
		},
		{
			name: "Fail",
			mockFunc: func(mock sqlmock.Sqlmock) {
				setupMockBD(mock, getRoleQuery, nil, fmt.Errorf("some error"), role.ExternalID)
			},
			id:      role.ExternalID,
			wantErr: true,
		},
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc(mock)
			q := New(db)
			_, err := q.GetRole(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Queries.GetRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestQueries_UpdateRole(t *testing.T) {
	ctx := context.Background()
	q := New(db)

	updateDate := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)
	role := createRole()
	rows := getRow(role, updateDate)

	tests := []struct {
		name    string
		arg     UpdateRoleParams
		wantErr bool
	}{
		{
			name: "Update Existing Role",
			arg: UpdateRoleParams{
				Name:        sql.NullString{String: role.Name, Valid: true},
				Description: sql.NullString{String: role.Description, Valid: true},
				ExternalID:  role.ExternalID,
				UpdateAt:    updateDate,
			},
			wantErr: false,
		},
		{
			name: "Invalid External ID",
			arg: UpdateRoleParams{
				Name:        sql.NullString{String: role.Name, Valid: true},
				Description: sql.NullString{String: role.Description, Valid: true},
				ExternalID:  "invalid-id",
				UpdateAt:    updateDate,
			},
			wantErr: true,
		},
		{
			name: "Empty Name",
			arg: UpdateRoleParams{
				Name:        sql.NullString{String: "", Valid: false},
				Description: sql.NullString{String: role.Description, Valid: true},
				ExternalID:  role.ExternalID,
				UpdateAt:    updateDate,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mocker.ExpectQuery(regexp.QuoteMeta(updateRoleQuery)).
				WithArgs(tt.arg.Name, tt.arg.Description, tt.arg.UpdateAt, tt.arg.ExternalID).
				WillReturnRows(rows)

			_, err := q.UpdateRole(ctx, tt.arg)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func setupMockBD(moc pgxmock.PgxPoolIface, query string, rows *pgxmock.Rows, err error, args ...any) {
	if rows != nil {
		moc.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args...).WillReturnRows(rows).WillReturnError(err)
	} else {
		moc.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args...).WillReturnError(err)
	}
}

func getRow(role *Role, updateDate time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"internal_id", "name", "description", "external_id", "created_at", "updated_at"}).
		AddRow(
			role.InternalID,
			role.Name,
			role.Description,
			role.ExternalID,
			updateDate,
			role.CreatedAt,
		)
}
