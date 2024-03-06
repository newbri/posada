package db

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/newbri/posadamissportia/db/util"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func createRandomUserWithRole(role string, isDeleted bool) *User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	if err != nil {
		return nil
	}

	t := time.Now()
	return &User{
		Username:          util.RandomOwner(),
		HashedPassword:    hashedPassword,
		FullName:          fmt.Sprintf("%s %s", util.RandomOwner(), util.RandomOwner()),
		Email:             util.RandomEmail(),
		PasswordChangedAt: t,
		CreatedAt:         t,
		Role:              createRandomRole(role),
		IsDeleted:         isDeleted,
		DeletedAt:         time.Time{},
	}
}

func createMultipleRandomUserWithRole(amountUserToCreate uint, role string, isDeleted bool) []*User {
	var users []*User
	for i := uint(0); i < amountUserToCreate; i++ {
		users = append(users, createRandomUserWithRole(role, isDeleted))
	}
	return users
}

func createRandomRole(roleType string) *Role {
	t := time.Now()
	return &Role{
		InternalID:  uuid.New(),
		Name:        roleType,
		Description: util.RandomString(16),
		ExternalID:  fmt.Sprintf("URE%d", util.RandomInt(101, 999)),
		UpdatedAt:   t,
		CreatedAt:   t,
	}
}

func createRole() *Role {
	return &Role{
		InternalID:  uuid.New(),
		Name:        util.RandomString(6),
		Description: util.RandomString(10),
		ExternalID:  util.RandomString(8),
		UpdatedAt:   time.Now(),
		CreatedAt:   time.Now(),
	}
}

func getMockedExpectedUserRows(user *User) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"username", "hashed_password", "full_name", "email", "password_changed_at", "users.created_at", "is_deleted", "deleted_at", "p.internal_id", "p.name", "p.description", "p.external_id", "p.created_at", "p.updated_at"}).
		AddRow(
			&user.Username,
			&user.HashedPassword,
			&user.FullName,
			&user.Email,
			&user.PasswordChangedAt,
			&user.CreatedAt,
			&user.IsDeleted,
			&user.DeletedAt,
			&user.Role.InternalID,
			&user.Role.Name,
			&user.Role.Description,
			&user.Role.ExternalID,
			&user.Role.CreatedAt,
			&user.Role.UpdatedAt,
		)
}

func getMockedExpectedCreateUserRows(user *User) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"username", "hashed_password", "full_name", "email", "password_changed_at", "created_at", "role_id", "is_deleted", "deleted_at"}).
		AddRow(
			&user.Username,
			&user.HashedPassword,
			&user.FullName,
			&user.Email,
			&user.PasswordChangedAt,
			&user.CreatedAt,
			&user.Role.InternalID,
			&user.IsDeleted,
			&user.DeletedAt,
		)
}

func getMockedExpectedUpdateUserRows(user *User) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"username", "hashed_password", "full_name", "email", "password_changed_at", "created_at", "role_id", "is_deleted", "deleted_at"}).
		AddRow(
			&user.Username,
			&user.HashedPassword,
			&user.FullName,
			&user.Email,
			&user.PasswordChangedAt,
			&user.CreatedAt,
			&user.Role.InternalID,
			&user.IsDeleted,
			&user.DeletedAt,
		)
}

func getMockedWrongExpectedUserRows(user *User) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"username", "hashed_password", "full_name", "email", "password_changed_at", "created_at", "role_id", "is_deleted"}).
		AddRow(
			&user.Username,
			&user.HashedPassword,
			&user.FullName,
			&user.Email,
			&user.PasswordChangedAt,
			&user.CreatedAt,
			&user.Role.InternalID,
			&user.IsDeleted,
		)
}

func getMockedExpectedRoleRows(role *Role) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"internal_id", "name", "description", "external_id", "updated_at", "created_at"}).
		AddRow(
			&role.InternalID,
			&role.Name,
			&role.Description,
			&role.ExternalID,
			&role.UpdatedAt,
			&role.CreatedAt,
		)
}

func getMultipleMockedExpectedUserRows(users []*User) *sqlmock.Rows {
	rowHeading := sqlmock.NewRows([]string{"username", "hashed_password", "full_name", "email", "password_changed_at", "created_at", "role_id", "is_deleted", "deleted_at"})
	for _, user := range users {
		rowHeading = rowHeading.AddRow(
			&user.Username,
			&user.HashedPassword,
			&user.FullName,
			&user.Email,
			&user.PasswordChangedAt,
			&user.CreatedAt,
			&user.Role.InternalID,
			&user.IsDeleted,
			&user.DeletedAt,
		)
	}
	return rowHeading
}

func getMultipleWrongMockedExpectedUserRows(users []*User) *sqlmock.Rows {
	rowHeading := sqlmock.NewRows([]string{"username", "hashed_password", "full_name", "email", "password_changed_at", "created_at", "role_id", "is_deleted"})
	for _, user := range users {
		rowHeading = rowHeading.AddRow(
			&user.Username,
			&user.HashedPassword,
			&user.FullName,
			&user.Email,
			&user.PasswordChangedAt,
			&user.CreatedAt,
			&user.Role.InternalID,
			&user.IsDeleted,
		)
	}
	return rowHeading
}

func createUserParams(user *User) *CreateUserParams {
	return &CreateUserParams{
		Username:       user.Username,
		HashedPassword: user.HashedPassword,
		FullName:       user.FullName,
		Email:          user.Email,
		RoleID:         user.Role.InternalID,
	}
}
