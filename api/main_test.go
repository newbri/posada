package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/newbri/posadamissportia/db"
	"github.com/newbri/posadamissportia/db/util"
	"github.com/newbri/posadamissportia/token"
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

func newTestServer(store db.Store, env string) *Server {
	appConfig := util.NewYAMLConfiguration("../app.yaml", env)

	maker, err := token.NewPasetoMaker(appConfig.GetConfig().TokenSymmetricKey)
	if err != nil {
		return nil
	}

	return NewServer(store, maker, appConfig)
}

func newServer(store db.Store, maker token.Maker, env string) *Server {
	appConfig := util.NewYAMLConfiguration("../app.yaml", env)
	return NewServer(store, maker, appConfig)
}

func createRandomUser(role string, isDeleted bool) *db.User {
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
		IsDeleted:         isDeleted,
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
	userToken, _, err := tokenMaker.CreateToken(username, role, duration)
	require.NoError(t, err)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, userToken)
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
}

func getAllRole() []*db.Role {
	var roles []*db.Role
	roles = append(roles, createRandomRole(db.RoleAdmin))
	roles = append(roles, createRandomRole(db.RoleVisitor))
	roles = append(roles, createRandomRole(db.RoleCustomer))
	return roles
}

func createSession(user *db.User, refreshToken string) *db.Session {
	t := time.Now()
	return &db.Session{
		ID:           uuid.New(),
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    "PostmanRuntime/7.36.0",
		ClientIp:     "::1",
		IsBlocked:    false,
		ExpiredAt:    t.Add(time.Hour),
		CreatedAt:    t,
	}
}
