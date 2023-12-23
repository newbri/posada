package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestQueries_UpdateRole(t *testing.T) {
	ctx := context.Background()
	db, _ := sql.Open("sqlite3", ":memory:")

	q := New(db)

	// Setting up test data
	roleParams := CreateRoleParams{
		ID:          uuid.New(),
		Name:        "Admin",
		Description: "Administrator role",
	}

	_, err := q.CreateRole(ctx, roleParams)
	require.NoError(t, err)

	tests := []struct {
		name    string
		arg     UpdateRoleParams
		wantErr bool
	}{
		{
			name: "Update Existing Role",
			arg: UpdateRoleParams{
				Name:        sql.NullString{String: "SuperAdmin", Valid: true},
				Description: sql.NullString{String: "Super Administrator role", Valid: true},
				ExternalID:  roleParams.ID.String(),
			},
			wantErr: false,
		},
		{
			name: "Invalid External ID",
			arg: UpdateRoleParams{
				Name:        sql.NullString{String: "User", Valid: true},
				Description: sql.NullString{String: "User role", Valid: true},
				ExternalID:  "invalid-id",
			},
			wantErr: true,
		},
		{
			name: "Empty Name",
			arg: UpdateRoleParams{
				Name:        sql.NullString{String: "", Valid: false},
				Description: sql.NullString{String: "User role", Valid: true},
				ExternalID:  roleParams.ID.String(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := q.UpdateRole(ctx, tt.arg)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}