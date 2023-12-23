package db

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestQueries_UpdateRole(t *testing.T) {
	ctx := context.Background()
	q := New(db)

	updateDate := time.Now()

	role := createRole()
	rows := sqlmock.NewRows([]string{"internal_id", "name", "description", "external_id", "updated_at", "created_at"}).
		AddRow(
			role.InternalID,
			role.Name,
			role.Description,
			role.ExternalID,
			updateDate,
			role.CreatedAt,
		)

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
