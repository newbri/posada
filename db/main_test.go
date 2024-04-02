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

func createSessionParams() *CreateSessionParams {
	return &CreateSessionParams{
		ID:           uuid.New(),
		Username:     util.RandomOwner(),
		RefreshToken: "v2.local.062BnLQ2SfzDwxpDeiHiF0Kv2nO4Ixfz47pTVKctf8Dln42rngnWuqDIb1VclXsDNJ4QWBYHqsVYriHjMnJ25sCHObv98yDFAt7UKgO1w6x9UUI7t_I6LWm6cnB-DOS4gzYWk9UXSzOKVgBTk5OPFiCNkQAorToPAJbIXEZmLSa4Niq1M6unXwoZwVP2HBLvOgBvBrL6PYHrwegeCXdq8Ce_izphetqzzHWRmqjq6J-3MSSeN0J4ZayZt9SN_Lv93zkuKajLTvpdPBDKY35VDkZel0wLzmXamnq8JBvpfixepJmtJCo-Ja-QyLat0qLuBXrVlYxR7kE0oFeBhrZ4KxazC4g3RTHYev8WfQM4HkUTtYpDZFD6AupeBjkp1q4vRQuWI5PtyDGT0l8aXFQwQdvuut4TCi5U_B0TghLFuEeUrmLylpYIQ_mJyXDu1YaIiVP5ZwCKkbERbb7hB5tETQ_1BMMZT1iPsBSleZ4Bam-x16B5fEJTs1czbXuLKw.bnVsbA",
		UserAgent:    "PostmanRuntime/7.36.3",
		ClientIp:     "::1",
		IsBlocked:    false,
		ExpiredAt:    time.Now().Add(time.Minute * 15),
		CreatedAt:    time.Now(),
	}
}

func getMockedExpectedCreateSession(session *Session) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "username", "refresh_token", "user_agent", "client_ip", "is_blocked", "expired_at", "created_at", "blocked_at"}).
		AddRow(
			&session.ID,           // id
			&session.Username,     // username
			&session.RefreshToken, // refresh_token
			&session.UserAgent,    // user_agent
			&session.ClientIp,     // client_ip
			&session.IsBlocked,    // is_blocked
			&session.ExpiredAt,    // expired_at
			&session.CreatedAt,    // created_at
			&session.BlockedAt,    // blocked_at
		)
}

func createSession(arg *CreateSessionParams) *Session {
	return &Session{
		ID:           arg.ID,
		Username:     arg.Username,
		RefreshToken: arg.RefreshToken,
		UserAgent:    arg.UserAgent,
		ClientIp:     arg.ClientIp,
		IsBlocked:    false,
		CreatedAt:    arg.CreatedAt,
		ExpiredAt:    arg.ExpiredAt,
		BlockedAt:    time.Time{},
	}
}

func getMockedExpectedCreateProperty(property *Property) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"internal_id", "external_id", "name", "address", "state", "city", "country", "postal_code", "phone", "email", "is_active", "expired_at", "created_at"}).
		AddRow(
			&property.InternalID, // internal_id
			&property.ExternalID, // external_id
			&property.Name,       // name
			&property.Address,    // address
			&property.State,      // state
			&property.City,       // city
			&property.Country,    // country
			&property.PostalCode, // postal_code
			&property.Phone,      // phone
			&property.Email,      // email
			&property.IsActive,   // is_active
			&property.ExpiredAt,  // expired_at
			&property.CreatedAt,  // created_at
		)
}

func createProperty(arg *CreatePropertyParams) *Property {
	return &Property{
		InternalID: uuid.New(),
		ExternalID: "PRO103",
		Name:       arg.Name,
		Address:    arg.Address,
		State:      arg.State,
		City:       arg.City,
		Country:    arg.Country,
		PostalCode: arg.PostalCode,
		Phone:      arg.Phone,
		Email:      arg.Email,
		IsActive:   false,
		ExpiredAt:  arg.ExpiredAt,
		CreatedAt:  arg.CreatedAt,
	}
}

func createPropertyParams() *CreatePropertyParams {
	return &CreatePropertyParams{
		Name:       util.RandomString(9),
		Address:    util.RandomString(9),
		City:       util.RandomString(9),
		State:      util.RandomString(9),
		Country:    util.RandomString(9),
		PostalCode: util.RandomString(6),
		Phone:      "438 830 7862",
		Email:      util.RandomEmail(),
		ExpiredAt:  time.Time{},
		CreatedAt:  time.Now(),
	}
}

func getMockedExpectedProperty(property *Property) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"internal_id", "external_id", "name", "address", "state", "city", "country", "postal_code", "phone", "email", "is_active", "expired_at", "created_at"}).
		AddRow(
			&property.InternalID, // internal_id
			&property.ExternalID, // external_id
			&property.Name,       // name
			&property.Address,    // address
			&property.State,      // state
			&property.City,       // city
			&property.Country,    // country
			&property.PostalCode, // postal_code
			&property.Phone,      // phone
			&property.Email,      // email
			&property.IsActive,   // is_active
			&property.ExpiredAt,  // expired_at
			&property.CreatedAt,  // created_at
		)
}
