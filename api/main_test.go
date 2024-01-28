package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/newbri/posadamissportia/db"
	"github.com/newbri/posadamissportia/db/util"
	"github.com/newbri/posadamissportia/token"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func newTestServer(store db.Store) *Server {
	config, err := util.LoadConfig("../app.yaml", "test")
	if err != nil {
		log.Fatal().Msg("cannot load config")
	}
	maker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil
	}

	return NewServer(store, maker, config)
}

func newServer(store db.Store, maker token.Maker, env string) *Server {
	config, err := util.LoadConfig("../app.yaml", env)
	if err != nil {
		log.Fatal().Msg("cannot load config")
	}

	return NewServer(store, maker, config)
}

func createRandomUser(role string) *db.User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	if err != nil {
		return nil
	}

	t := time.Now()
	return &db.User{
		Username:          util.RandomOwner(),
		HashedPassword:    hashedPassword,
		FullName:          fmt.Sprintf("%s %s", util.RandomOwner(), util.RandomOwner()),
		Email:             util.RandomEmail(),
		PasswordChangedAt: t,
		CreatedAt:         t,
		Role:              createRandomRole(role),
		IsDeleted:         false,
		DeletedAt:         time.Time{},
	}
}

func createRandomRole(roleType string) *db.Role {
	t := time.Now()
	return &db.Role{
		InternalID:  uuid.New(),
		Name:        roleType,
		Description: util.RandomString(16),
		ExternalID:  fmt.Sprintf("URE%d", util.RandomInt(101, 999)),
		UpdatedAt:   t,
		CreatedAt:   t,
	}
}

func createRandomUserAndPassword() (*db.User, string) {
	password := util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	if err != nil {
		return nil, ""
	}

	t := time.Now()
	return &db.User{
		Username:          util.RandomOwner(),
		HashedPassword:    hashedPassword,
		FullName:          fmt.Sprintf("%s %s", util.RandomOwner(), util.RandomOwner()),
		Email:             util.RandomEmail(),
		PasswordChangedAt: t,
		CreatedAt:         t,
		Role:              createRandomRole(db.RoleAdmin),
	}, password
}

func createToken(symmetricKey string, username string, role *db.Role, duration time.Duration) (string, *token.Payload, error) {
	tokenMaker, err := token.NewPasetoMaker(symmetricKey)
	if err != nil {
		return "", nil, err
	}
	return tokenMaker.CreateToken(username, role, duration)
}

func addAuthorization(t *testing.T, request *http.Request, tokenMaker token.Maker, authorizationType string, username string, role *db.Role, duration time.Duration) {
	userToken, payload, err := tokenMaker.CreateToken(username, role, duration)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, userToken)
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
}

func testGetAllRole() []*db.Role {
	var roles []*db.Role
	roles = append(roles, createRandomRole(db.RoleAdmin))
	roles = append(roles, createRandomRole(db.RoleVisitor))
	roles = append(roles, createRandomRole(db.RoleCustomer))
	return roles
}
