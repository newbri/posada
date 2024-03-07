package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/newbri/posadamissportia/configuration"
	"github.com/newbri/posadamissportia/db"
	"github.com/newbri/posadamissportia/db/mocker"
	"github.com/newbri/posadamissportia/db/util"
	"github.com/newbri/posadamissportia/token"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x any) bool {
	args, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(e.password, args.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = args.HashedPassword
	return reflect.DeepEqual(e.arg, args)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

type eqUpdateUserParamsMatcher struct {
	arg      db.UpdateUserParams
	password string
}

func (e eqUpdateUserParamsMatcher) Matches(x any) bool {
	args, ok := x.(db.UpdateUserParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(e.password, args.HashedPassword.String)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = args.HashedPassword
	return true
}

func (e eqUpdateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqUpdateUserParams(arg db.UpdateUserParams, password string) gomock.Matcher {
	return eqUpdateUserParamsMatcher{arg, password}
}

func TestCreateUser(t *testing.T) {
	password := "lexy84"
	longPassword := util.RandomString(73)
	expectedUser := createRandomUser(db.RoleAdmin, false)

	testCases := []struct {
		name     string
		env      string
		body     gin.H
		mock     func(server *Server)
		response func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			env:  "test",
			body: gin.H{
				"username":  expectedUser.Username,
				"password":  password,
				"full_name": expectedUser.FullName,
				"email":     expectedUser.Email,
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				role := createRandomRole(db.RoleCustomer)
				querier.
					On("GetRoleByName", mock.Anything, role.Name).
					Return(role, nil)

				querier.
					On("CreateUser", mock.Anything, mock.Anything).
					Return(expectedUser, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)

				var request userResponse
				err = json.Unmarshal(data, &request)
				require.NoError(t, err)

				require.Equal(t, expectedUser.Username, request.Username)
				require.Equal(t, expectedUser.FullName, request.FullName)
				require.Equal(t, expectedUser.Email, request.Email)
				require.Equal(t, expectedUser.PasswordChangedAt.Unix(), request.PasswordChangedAt.Unix())
				require.Equal(t, expectedUser.CreatedAt.Unix(), request.CreatedAt.Unix())
			},
		},
		{
			name: "StatusBadRequest",
			env:  "test",
			body: gin.H{
				"username":  expectedUser.Username,
				"password":  password,
				"full_name": expectedUser.FullName,
				"email1":    expectedUser.Email,
			},
			mock: func(server *Server) {
				arg := db.CreateUserParams{
					Username: expectedUser.Username,
					FullName: expectedUser.FullName,
					Email:    expectedUser.Email,
				}

				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("CreateUser", mock.Anything, EqCreateUserParams(arg, password)).
					Return(expectedUser, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadPassword",
			env:  "test",
			body: gin.H{
				"username":  expectedUser.Username,
				"password":  longPassword,
				"full_name": expectedUser.FullName,
				"email":     expectedUser.Email,
			},
			mock: func(server *Server) {
				arg := db.CreateUserParams{
					Username: expectedUser.Username,
					FullName: expectedUser.FullName,
					Email:    expectedUser.Email,
				}

				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("CreateUser", mock.Anything, EqCreateUserParams(arg, longPassword)).
					Return(nil, fmt.Errorf("error creating user"))
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "DuplicateUsername",
			env:  "test",
			body: gin.H{
				"username":  expectedUser.Username,
				"password":  password,
				"full_name": expectedUser.FullName,
				"email":     expectedUser.Email,
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				role := createRandomRole(db.RoleCustomer)

				querier.
					On("GetRoleByName", mock.Anything, mock.Anything).
					Return(role, nil)

				querier.
					On("CreateUser", mock.Anything, mock.Anything).
					Return(nil, &pq.Error{Code: "23505"})
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InternalError",
			env:  "test",
			body: gin.H{
				"username":  expectedUser.Username,
				"password":  password,
				"full_name": expectedUser.FullName,
				"email":     expectedUser.Email,
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				role := createRandomRole(db.RoleCustomer)
				querier.
					On("GetRoleByName", mock.Anything, mock.Anything).
					Return(role, nil)

				querier.
					On("CreateUser", mock.Anything, mock.Anything).
					Return(nil, sql.ErrConnDone)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "RoleDoesNotExist",
			env:  "test",
			body: gin.H{
				"username":  expectedUser.Username,
				"password":  password,
				"full_name": expectedUser.FullName,
				"email":     expectedUser.Email,
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetRoleByName", mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("error exist"))
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			querier := new(mocker.TestMocker)
			server := newTestServer(querier, tc.env)
			tc.mock(server)

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/users"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			server.router.ServeHTTP(recorder, request)
			tc.response(recorder)
		})
	}
}

func TestGetUser(t *testing.T) {
	adminUser := createRandomUser(db.RoleAdmin, false)

	testCases := []struct {
		name     string
		username string
		env      string
		mock     func(server *Server)
		response func(recorder *httptest.ResponseRecorder)
		auth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			name:     "OK",
			env:      "test",
			username: adminUser.Username,
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Return(adminUser, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)

				var request userResponse
				err = json.Unmarshal(data, &request)
				require.NoError(t, err)

				require.Equal(t, adminUser.Username, request.Username)
				require.Equal(t, adminUser.FullName, request.FullName)
				require.Equal(t, adminUser.Email, request.Email)
				require.Equal(t, adminUser.PasswordChangedAt.Unix(), request.PasswordChangedAt.Unix())
				require.Equal(t, adminUser.CreatedAt.Unix(), request.CreatedAt.Unix())
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					adminUser.Username,
					adminUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusBadRequest",
			env:      "test",
			username: "-@",
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Return(adminUser, nil)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(0)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					adminUser.Username,
					adminUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusNotFound",
			env:      "test",
			username: adminUser.Username,
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(adminUser, nil)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(nil, sql.ErrNoRows)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					adminUser.Username,
					adminUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusInternalServerError",
			env:      "test",
			username: adminUser.Username,
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(adminUser, nil)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					adminUser.Username,
					adminUser.Role,
					time.Minute,
				)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			querier := new(mocker.TestMocker)
			server := newTestServer(querier, tc.env)
			tc.mock(server)

			url := fmt.Sprintf("/api/v1/auth/admin/users/%s", tc.username)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.auth(t, request, server.tokenMaker)
			recorder := httptest.NewRecorder()
			server.router.ServeHTTP(recorder, request)
			tc.response(recorder)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	password := "lexy84"
	longPassword := util.RandomString(73)
	customerUser := createRandomUser(db.RoleCustomer, false)

	testCases := []struct {
		name     string
		env      string
		body     gin.H
		username string
		mock     func(server *Server)
		response func(recorder *httptest.ResponseRecorder)
		auth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			name: "OK",
			env:  "test",
			body: gin.H{
				"username": customerUser.Username,
				"email":    customerUser.Email,
			},
			username: customerUser.Username,
			mock: func(server *Server) {
				args := db.UpdateUserParams{
					Username: customerUser.Username,
					Email: sql.NullString{
						String: customerUser.Email,
						Valid:  true,
					},
				}

				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(customerUser, nil)

				querier.
					On("UpdateUser", mock.AnythingOfType("*gin.Context"), args).
					Times(1).
					Return(customerUser, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)

				var request userResponse
				err = json.Unmarshal(data, &request)
				require.NoError(t, err)

				require.Equal(t, customerUser.Username, request.Username)
				require.Equal(t, customerUser.FullName, request.FullName)
				require.Equal(t, customerUser.Email, request.Email)
				require.Equal(t, customerUser.PasswordChangedAt.Unix(), request.PasswordChangedAt.Unix())
				require.Equal(t, customerUser.CreatedAt.Unix(), request.CreatedAt.Unix())
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					customerUser.Username,
					customerUser.Role,
					time.Minute,
				)
			},
		},
		{
			name: "OK Changing Password",
			env:  "test",
			body: gin.H{
				"username":  customerUser.Username,
				"password":  password,
				"full_name": customerUser.FullName,
			},
			username: customerUser.Username,
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(customerUser, nil)

				passwordMatcher := mock.MatchedBy(func(args db.UpdateUserParams) bool {
					return util.CheckPassword(password, args.HashedPassword.String) == nil
				})

				querier.
					On("UpdateUser", mock.Anything, passwordMatcher).
					Times(1).
					Return(customerUser, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)

				var request userResponse
				err = json.Unmarshal(data, &request)
				require.NoError(t, err)

				require.Equal(t, customerUser.Username, request.Username)
				require.Equal(t, customerUser.FullName, request.FullName)
				require.Equal(t, customerUser.Email, request.Email)
				require.Equal(t, customerUser.PasswordChangedAt.Unix(), request.PasswordChangedAt.Unix())
				require.Equal(t, customerUser.CreatedAt.Unix(), request.CreatedAt.Unix())
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					customerUser.Username,
					customerUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusBadRequest",
			env:      "test",
			username: "-@",
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(customerUser, nil)

				querier.
					On("UpdateUser", mock.Anything, mock.Anything).
					Times(0)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					customerUser.Username,
					customerUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusInternalServerError with long password",
			env:      "test",
			username: customerUser.Username,
			body: gin.H{
				"username":  customerUser.Username,
				"password":  longPassword,
				"full_name": customerUser.FullName,
			},
			mock: func(server *Server) {
				args := db.UpdateUserParams{
					Username: customerUser.Username,
					FullName: sql.NullString{
						String: customerUser.FullName,
						Valid:  true,
					},
				}

				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(customerUser, nil)

				querier.
					On("UpdateUser", mock.Anything, EqUpdateUserParams(args, longPassword)).
					Times(0)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					customerUser.Username,
					customerUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusNotFound",
			env:      "test",
			username: customerUser.Username,
			body: gin.H{
				"username":  customerUser.Username,
				"password":  password,
				"full_name": customerUser.FullName,
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(customerUser, nil)

				passwordMatcher := mock.MatchedBy(func(args db.UpdateUserParams) bool {
					return util.CheckPassword(password, args.HashedPassword.String) == nil
				})

				querier.
					On("UpdateUser", mock.Anything, passwordMatcher).
					Times(1).
					Return(nil, sql.ErrNoRows)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					customerUser.Username,
					customerUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusInternalServerError",
			env:      "test",
			username: customerUser.Username,
			body: gin.H{
				"username":  customerUser.Username,
				"password":  password,
				"full_name": customerUser.FullName,
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(customerUser, nil)

				passwordMatcher := mock.MatchedBy(func(args db.UpdateUserParams) bool {
					return util.CheckPassword(password, args.HashedPassword.String) == nil
				})

				querier.
					On("UpdateUser", mock.Anything, passwordMatcher).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					customerUser.Username,
					customerUser.Role,
					time.Minute,
				)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			querier := new(mocker.TestMocker)
			server := newTestServer(querier, tc.env)
			tc.mock(server)

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/auth/customer/users"
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.auth(t, request, server.tokenMaker)
			recorder := httptest.NewRecorder()
			server.router.ServeHTTP(recorder, request)
			tc.response(recorder)
		})
	}
}

func TestDeleteUser(t *testing.T) {
	adminUser := createRandomUser(db.RoleAdmin, false)

	testCases := []struct {
		name     string
		env      string
		username string
		mock     func(server *Server)
		response func(recorder *httptest.ResponseRecorder)
		auth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			name:     "OK",
			env:      "test",
			username: adminUser.Username,
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(adminUser, nil)

				querier.
					On("DeleteUser", mock.Anything, mock.Anything, mock.Anything).
					Times(1).
					Return(adminUser, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)

				var request userResponse
				err = json.Unmarshal(data, &request)
				require.NoError(t, err)

				require.Equal(t, adminUser.Username, request.Username)
				require.Equal(t, adminUser.FullName, request.FullName)
				require.Equal(t, adminUser.Email, request.Email)
				require.Equal(t, adminUser.PasswordChangedAt.Unix(), request.PasswordChangedAt.Unix())
				require.Equal(t, adminUser.CreatedAt.Unix(), request.CreatedAt.Unix())
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					adminUser.Username,
					adminUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusBadRequest",
			env:      "test",
			username: "-@",
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(adminUser, nil)

				querier.
					On("DeleteUser", mock.Anything, mock.Anything, mock.Anything).
					Times(0)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					adminUser.Username,
					adminUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusNotFound",
			env:      "test",
			username: adminUser.Username,
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(adminUser, nil)

				querier.
					On("DeleteUser", mock.Anything, mock.Anything, mock.Anything).
					Times(1).
					Return(nil, sql.ErrNoRows)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					adminUser.Username,
					adminUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusInternalServerError",
			env:      "test",
			username: adminUser.Username,
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(adminUser, nil)

				querier.
					On("DeleteUser", mock.Anything, mock.Anything, mock.Anything).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					adminUser.Username,
					adminUser.Role,
					time.Minute,
				)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			querier := new(mocker.TestMocker)
			server := newTestServer(querier, tc.env)
			tc.mock(server)

			url := fmt.Sprintf("/api/v1/auth/admin/users/%s", tc.username)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.auth(t, request, server.tokenMaker)
			recorder := httptest.NewRecorder()
			server.router.ServeHTTP(recorder, request)
			tc.response(recorder)
		})
	}
}

func TestLoginUser(t *testing.T) {
	expectedUser, password := createRandomUserAndPassword()

	testCases := []struct {
		name     string
		env      string
		body     gin.H
		mock     func(server *Server)
		response func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			env:  "test",
			body: gin.H{
				"username": expectedUser.Username,
				"password": password,
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Return(expectedUser, nil)

				accessToken, accessPayload, err := createToken(
					server.config.GetConfig().TokenSymmetricKey,
					expectedUser.Username,
					expectedUser.Role,
					server.config.GetConfig().AccessTokenDuration,
				)
				require.NoError(t, err)

				maker, ok := server.tokenMaker.(*mocker.TestMocker)
				require.True(t, ok)

				maker.
					On("CreateToken",
						expectedUser.Username,
						expectedUser.Role,
						server.config.GetConfig().AccessTokenDuration,
					).
					Return(accessToken, accessPayload, nil)

				refreshToken, refreshPayload, err := createToken(
					server.config.GetConfig().TokenSymmetricKey,
					expectedUser.Username,
					expectedUser.Role,
					server.config.GetConfig().RefreshTokenDuration,
				)
				require.NoError(t, err)

				maker.
					On("CreateToken",
						expectedUser.Username,
						expectedUser.Role,
						server.config.GetConfig().RefreshTokenDuration,
					).
					Return(refreshToken, refreshPayload, nil)

				session := createSession(expectedUser, refreshToken)
				querier.
					On("CreateSession", mock.Anything, mock.Anything).
					Return(session, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			env:  "test",
			name: "UserNotFound",
			body: gin.H{
				"username": "NotFound",
				"password": password,
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(nil, sql.ErrNoRows)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			env:  "test",
			name: "IncorrectPassword",
			body: gin.H{
				"username": expectedUser.Username,
				"password": "incorrect",
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(expectedUser, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			env:  "test",
			name: "InternalError",
			body: gin.H{
				"username": expectedUser.Username,
				"password": password,
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			env:  "test",
			name: "StatusBadRequest",
			body: gin.H{
				"username":  "NotFound",
				"password1": password,
			},
			mock: func(server *Server) {

			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			env:  "test",
			name: "InvalidUsername",
			body: gin.H{
				"username": expectedUser.Username,
				"password": password,
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(nil, sql.ErrNoRows)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			env:  "test",
			name: "AccessTokenError",
			body: gin.H{
				"username": expectedUser.Username,
				"password": password,
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, expectedUser.Username).
					Return(expectedUser, nil).Once()

				maker, ok := server.tokenMaker.(*mocker.TestMocker)
				require.True(t, ok)

				err := errors.New("failed to create chacha20poly1305 cipher")
				maker.
					On(
						"CreateToken",
						expectedUser.Username,
						expectedUser.Role,
						server.config.GetConfig().AccessTokenDuration,
					).
					Return("", nil, err)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			env:  "test",
			name: "RefreshTokenError",
			body: gin.H{
				"username": expectedUser.Username,
				"password": password,
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Return(expectedUser, nil)

				accessToken, accessPayload, err := createToken(
					server.config.GetConfig().TokenSymmetricKey,
					expectedUser.Username,
					expectedUser.Role,
					server.config.GetConfig().AccessTokenDuration,
				)

				maker, ok := server.tokenMaker.(*mocker.TestMocker)
				require.True(t, ok)

				maker.
					On("CreateToken",
						expectedUser.Username,
						expectedUser.Role,
						server.config.GetConfig().AccessTokenDuration,
					).
					Return(accessToken, accessPayload, err)

				err = errors.New("failed to create chacha20poly1305 cipher")
				maker.
					On("CreateToken",
						expectedUser.Username,
						expectedUser.Role,
						server.config.GetConfig().RefreshTokenDuration,
					).
					Return("", nil, err)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			env:  "test",
			name: "SessionError",
			body: gin.H{
				"username": expectedUser.Username,
				"password": password,
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(expectedUser, nil)

				querier.
					On("CreateSession", mock.Anything, mock.Anything).
					Times(1).
					Return(nil, sql.ErrNoRows)

				accessToken, accessPayload, err := createToken(
					server.config.GetConfig().TokenSymmetricKey,
					expectedUser.Username,
					expectedUser.Role,
					server.config.GetConfig().AccessTokenDuration,
				)

				querier.
					On("CreateToken", mock.Anything, mock.Anything, mock.Anything).
					Times(1).
					Return(accessToken, accessPayload, err)

				refreshToken, refreshPayload, err := createToken(
					server.config.GetConfig().TokenSymmetricKey,
					expectedUser.Username,
					expectedUser.Role,
					server.config.GetConfig().RefreshTokenDuration,
				)
				querier.
					On("CreateToken", mock.Anything, mock.Anything, mock.Anything).
					Times(1).
					Return(refreshToken, refreshPayload, err)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			querier := new(mocker.TestMocker)
			config := configuration.NewYAMLConfiguration("../app.yaml", tc.env)
			server := newServerWithConfigurator(querier, querier, config)
			tc.mock(server)

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/users/login"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			server.router.ServeHTTP(recorder, request)
			tc.response(recorder)

			querier.AssertExpectations(t)
		})
	}
}

func TestUserInfo(t *testing.T) {
	customerUser := createRandomUser(db.RoleCustomer, false)

	testCases := []struct {
		name     string
		env      string
		username string
		mock     func(server *Server)
		response func(recorder *httptest.ResponseRecorder)
		auth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			name:     "OK",
			env:      "test",
			username: customerUser.Username,
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.AnythingOfType("*gin.Context"), customerUser.Username).
					Return(customerUser, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)

				var request userResponse
				err = json.Unmarshal(data, &request)
				require.NoError(t, err)

				require.Equal(t, customerUser.Username, request.Username)
				require.Equal(t, customerUser.FullName, request.FullName)
				require.Equal(t, customerUser.Email, request.Email)
				require.Equal(t, customerUser.PasswordChangedAt.Unix(), request.PasswordChangedAt.Unix())
				require.Equal(t, customerUser.CreatedAt.Unix(), request.CreatedAt.Unix())
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					customerUser.Username,
					customerUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusNotFound",
			env:      "test",
			username: customerUser.Username,
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(customerUser, nil)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(nil, sql.ErrNoRows)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					customerUser.Username,
					customerUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusInternalServerError",
			env:      "test",
			username: customerUser.Username,
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, customerUser.Username).
					Times(1).
					Return(customerUser, nil)

				querier.
					On("GetUser", mock.Anything, customerUser.Username).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					customerUser.Username,
					customerUser.Role,
					time.Minute,
				)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			querier := new(mocker.TestMocker)
			server := newTestServer(querier, tc.env)
			tc.mock(server)

			url := "/api/v1/auth/customer/users/info"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.auth(t, request, server.tokenMaker)
			recorder := httptest.NewRecorder()
			server.router.ServeHTTP(recorder, request)
			tc.response(recorder)
		})
	}
}

func TestGetAllCustomer(t *testing.T) {
	adminUser := createRandomUser(db.RoleAdmin, false)
	var allCustomer []*db.User
	for i := 0; i < 6; i++ {
		allCustomer = append(allCustomer, createRandomUser(db.RoleCustomer, false))
	}

	testCases := []struct {
		name     string
		env      string
		body     gin.H
		mock     func(server *Server)
		response func(recorder *httptest.ResponseRecorder)
		auth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			name: "OK",
			env:  "test",
			body: gin.H{
				"offset": 0,
				"limit":  6,
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(adminUser, nil)

				querier.
					On("GetAllCustomer", mock.Anything, mock.Anything).
					Times(1).
					Return(allCustomer, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)

				var request []*userResponse
				err = json.Unmarshal(data, &request)
				require.NoError(t, err)

				for i, user := range allCustomer {
					expectedUser := &userResponse{
						Username:          user.Username,
						FullName:          user.FullName,
						Email:             user.Email,
						PasswordChangedAt: user.PasswordChangedAt,
						CreatedAt:         user.CreatedAt,
						Role:              user.Role,
					}
					require.Equal(t, expectedUser.Username, request[i].Username)
					require.Equal(t, expectedUser.FullName, request[i].FullName)
					require.Equal(t, expectedUser.Email, request[i].Email)
					require.Equal(t, expectedUser.PasswordChangedAt.Unix(), request[i].PasswordChangedAt.Unix())
					require.Equal(t, expectedUser.CreatedAt.Unix(), request[i].CreatedAt.Unix())
				}
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					adminUser.Username,
					adminUser.Role,
					time.Minute,
				)
			},
		},
		{
			name: "BadRequest",
			env:  "test",
			body: gin.H{
				"offset": 0,
				"limit1": 6,
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(adminUser, nil)

				querier.
					On("GetAllCustomer", mock.Anything, mock.Anything).
					Times(0)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					adminUser.Username,
					adminUser.Role,
					time.Minute,
				)
			},
		},
		{
			name: "ErrNoRows",
			env:  "test",
			body: gin.H{
				"offset": 0,
				"limit":  6,
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(adminUser, nil)

				querier.
					On("GetAllCustomer", mock.Anything, mock.Anything).
					Times(1).
					Return(nil, sql.ErrNoRows)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					adminUser.Username,
					adminUser.Role,
					time.Minute,
				)
			},
		},
		{
			name: "InternalServerError",
			env:  "test",
			body: gin.H{
				"offset": 0,
				"limit":  6,
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(adminUser, nil)

				querier.
					On("GetAllCustomer", mock.Anything, mock.Anything).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					adminUser.Username,
					adminUser.Role,
					time.Minute,
				)
			},
		},
		{
			name: "NoCustomerFound",
			env:  "test",
			body: gin.H{
				"offset": 0,
				"limit":  6,
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(adminUser, nil)

				querier.
					On("GetAllCustomer", mock.Anything, mock.Anything).
					Times(1).
					Return(nil, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					adminUser.Username,
					adminUser.Role,
					time.Minute,
				)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			querier := new(mocker.TestMocker)
			server := newTestServer(querier, tc.env)
			tc.mock(server)

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/auth/admin/users/all/customer"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.auth(t, request, server.tokenMaker)
			recorder := httptest.NewRecorder()
			server.router.ServeHTTP(recorder, request)
			tc.response(recorder)
		})
	}
}

func TestGetAllAdmin(t *testing.T) {
	superUser := createRandomUser(db.RoleSuperUser, false)
	var allAdmin []*db.User
	for i := 0; i < 6; i++ {
		allAdmin = append(allAdmin, createRandomUser(db.RoleAdmin, false))
	}

	testCases := []struct {
		name     string
		env      string
		body     gin.H
		mock     func(server *Server)
		response func(recorder *httptest.ResponseRecorder)
		auth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			name: "OK",
			env:  "test",
			body: gin.H{
				"offset": 0,
				"limit":  6,
			},
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(superUser, nil)

				querier.
					On("GetAllAdmin", mock.Anything, mock.Anything).
					Times(1).
					Return(allAdmin, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)

				var request []*userResponse
				err = json.Unmarshal(data, &request)
				require.NoError(t, err)

				for i, user := range allAdmin {
					expectedUser := &userResponse{
						Username:          user.Username,
						FullName:          user.FullName,
						Email:             user.Email,
						PasswordChangedAt: user.PasswordChangedAt,
						CreatedAt:         user.CreatedAt,
						Role:              user.Role,
					}
					require.Equal(t, expectedUser.Username, request[i].Username)
					require.Equal(t, expectedUser.FullName, request[i].FullName)
					require.Equal(t, expectedUser.Email, request[i].Email)
					require.Equal(t, expectedUser.PasswordChangedAt.Unix(), request[i].PasswordChangedAt.Unix())
					require.Equal(t, expectedUser.CreatedAt.Unix(), request[i].CreatedAt.Unix())
				}
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					superUser.Username,
					superUser.Role,
					time.Minute,
				)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			querier := new(mocker.TestMocker)
			server := newTestServer(querier, tc.env)
			tc.mock(server)

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/auth/su/users/all/admin"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.auth(t, request, server.tokenMaker)
			recorder := httptest.NewRecorder()
			server.router.ServeHTTP(recorder, request)
			tc.response(recorder)
		})
	}
}
