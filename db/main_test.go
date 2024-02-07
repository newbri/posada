package db

import (
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/newbri/posadamissportia/db/util"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

var mocker sqlmock.Sqlmock
var db *sql.DB

func TestMain(m *testing.M) {
	var err error
	db, mocker, err = sqlmock.New()
	if err != nil {
		log.Fatal().Msgf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

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

func createRandomUser(t *testing.T) User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	return User{
		Username:          util.RandomOwner(),
		HashedPassword:    hashedPassword,
		FullName:          util.RandomOwner(),
		Email:             util.RandomEmail(),
		PasswordChangedAt: time.Now(),
		CreatedAt:         time.Now(),
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
