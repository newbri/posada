package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/newbri/posadamissportia/db"
	mockdb "github.com/newbri/posadamissportia/db/mock"
	"github.com/newbri/posadamissportia/db/util"
	"github.com/newbri/posadamissportia/token"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
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
	expectedUser := createRandomUser()

	testCases := []struct {
		name          string
		env           string
		body          gin.H
		buildStubs    func(server *Server)
		checkResponse func(recorder *httptest.ResponseRecorder)
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
			buildStubs: func(server *Server) {
				arg := db.CreateUserParams{
					RoleID:   uuid.MustParse("018cb346-945e-77d3-87b3-181d1b50a382"),
					Username: expectedUser.Username,
					FullName: expectedUser.FullName,
					Email:    expectedUser.Email,
				}

				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(1).
					Return(expectedUser, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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
			buildStubs: func(server *Server) {
				arg := db.CreateUserParams{
					Username: expectedUser.Username,
					FullName: expectedUser.FullName,
					Email:    expectedUser.Email,
				}

				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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
			buildStubs: func(server *Server) {
				arg := db.CreateUserParams{
					Username: expectedUser.Username,
					FullName: expectedUser.FullName,
					Email:    expectedUser.Email,
				}

				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, longPassword)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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
			buildStubs: func(server *Server) {
				arg := db.CreateUserParams{
					RoleID:   uuid.MustParse("018cb346-945e-77d3-87b3-181d1b50a382"),
					Username: expectedUser.Username,
					FullName: expectedUser.FullName,
					Email:    expectedUser.Email,
				}

				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(1).
					Return(nil, &pq.Error{Code: "23505"})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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
			buildStubs: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			server := newServer(store, nil, tc.env)
			tc.buildStubs(server)

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/users"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetUser(t *testing.T) {
	expectedUser := createRandomUser()

	testCases := []struct {
		name          string
		username      string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			name:     "OK",
			username: expectedUser.Username,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), expectedUser.Username).
					Times(1).
					Return(expectedUser, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					expectedUser.Username,
					expectedUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusBadRequest",
			username: "-@",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					expectedUser.Username,
					expectedUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusNotFound",
			username: expectedUser.Username,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					expectedUser.Username,
					expectedUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusInternalServerError",
			username: expectedUser.Username,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					expectedUser.Username,
					expectedUser.Role,
					time.Minute,
				)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/auth/admin/users/%s", tc.username)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	password := "lexy84"
	longPassword := util.RandomString(73)
	expectedUser := createRandomUser()

	testCases := []struct {
		name          string
		body          gin.H
		username      string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			name: "OK",
			body: gin.H{
				"username": expectedUser.Username,
				"email":    expectedUser.Email,
			},
			username: expectedUser.Username,
			buildStubs: func(store *mockdb.MockStore) {
				args := db.UpdateUserParams{
					Username: expectedUser.Username,
					Email: sql.NullString{
						String: expectedUser.Email,
						Valid:  true,
					},
				}

				store.EXPECT().
					UpdateUser(gomock.Any(), args).
					Times(1).
					Return(expectedUser, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					expectedUser.Username,
					expectedUser.Role,
					time.Minute,
				)
			},
		},
		{
			name: "OK Changing Password",
			body: gin.H{
				"username":  expectedUser.Username,
				"password":  password,
				"full_name": expectedUser.FullName,
			},
			username: expectedUser.Username,
			buildStubs: func(store *mockdb.MockStore) {
				args := db.UpdateUserParams{
					Username: expectedUser.Username,
					FullName: sql.NullString{
						String: expectedUser.FullName,
						Valid:  true,
					},
				}

				store.EXPECT().
					UpdateUser(gomock.Any(), EqUpdateUserParams(args, password)).
					Times(1).
					Return(expectedUser, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					expectedUser.Username,
					expectedUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusBadRequest",
			username: "-@",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					expectedUser.Username,
					expectedUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusInternalServerError with long password",
			username: expectedUser.Username,
			body: gin.H{
				"username":  expectedUser.Username,
				"password":  longPassword,
				"full_name": expectedUser.FullName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				args := db.UpdateUserParams{
					Username: expectedUser.Username,
					FullName: sql.NullString{
						String: expectedUser.FullName,
						Valid:  true,
					},
				}
				store.EXPECT().
					UpdateUser(gomock.Any(), EqUpdateUserParams(args, longPassword)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					expectedUser.Username,
					expectedUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusNotFound",
			username: expectedUser.Username,
			body: gin.H{
				"username":  expectedUser.Username,
				"password":  password,
				"full_name": expectedUser.FullName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				args := db.UpdateUserParams{
					Username: expectedUser.Username,
					FullName: sql.NullString{
						String: expectedUser.FullName,
						Valid:  true,
					},
				}
				store.EXPECT().
					UpdateUser(gomock.Any(), EqUpdateUserParams(args, password)).
					Times(1).
					Return(nil, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					expectedUser.Username,
					expectedUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusInternalServerError",
			username: expectedUser.Username,
			body: gin.H{
				"username":  expectedUser.Username,
				"password":  password,
				"full_name": expectedUser.FullName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				args := db.UpdateUserParams{
					Username: expectedUser.Username,
					FullName: sql.NullString{
						String: expectedUser.FullName,
						Valid:  true,
					},
				}
				store.EXPECT().
					UpdateUser(gomock.Any(), EqUpdateUserParams(args, password)).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					expectedUser.Username,
					expectedUser.Role,
					time.Minute,
				)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/auth/users"
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteUser(t *testing.T) {
	expectedUser := createRandomUser()

	testCases := []struct {
		name          string
		username      string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			name:     "OK",
			username: expectedUser.Username,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteUser(gomock.Any(), expectedUser.Username).
					Times(1).
					Return(expectedUser, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					expectedUser.Username,
					expectedUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusBadRequest",
			username: "-@",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteUser(gomock.Any(), expectedUser.Username).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					expectedUser.Username,
					expectedUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusNotFound",
			username: expectedUser.Username,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteUser(gomock.Any(), expectedUser.Username).
					Times(1).
					Return(nil, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					expectedUser.Username,
					expectedUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusInternalServerError",
			username: expectedUser.Username,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteUser(gomock.Any(), expectedUser.Username).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					expectedUser.Username,
					expectedUser.Role,
					time.Minute,
				)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/auth/admin/users/%s", tc.username)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestLoginUser(t *testing.T) {
	expectedUser, password := createRandomUserAndPassword()

	testCases := []struct {
		name          string
		env           string
		body          gin.H
		buildStubs    func(server *Server)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			env:  "test",
			body: gin.H{
				"username": expectedUser.Username,
				"password": password,
			},
			buildStubs: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(expectedUser, nil)

				accessToken, accessPayload, err := createToken(
					server.config.TokenSymmetricKey,
					expectedUser.Username,
					expectedUser.Role,
					server.config.AccessTokenDuration,
				)
				require.NoError(t, err)

				maker, ok := server.tokenMaker.(*mockdb.MockMaker)
				require.True(t, ok)

				maker.
					EXPECT().
					CreateToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(accessToken, accessPayload, nil).
					AnyTimes()

				refreshToken, refreshPayload, err := createToken(
					server.config.TokenSymmetricKey,
					expectedUser.Username,
					expectedUser.Role,
					server.config.RefreshTokenDuration,
				)
				require.NoError(t, err)
				maker.
					EXPECT().
					CreateToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(2).
					Return(refreshToken, refreshPayload, nil).
					AnyTimes()

				session := &db.Session{
					ID:           uuid.New(),
					Username:     expectedUser.Username,
					RefreshToken: refreshToken,
					UserAgent:    "PostmanRuntime/7.36.0",
					ClientIp:     "::1",
					IsBlocked:    false,
					ExpiredAt:    time.Now().Add(time.Hour),
					CreatedAt:    time.Now(),
				}

				store.
					EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(session, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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
			buildStubs: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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
			buildStubs: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					GetUser(gomock.Any(), gomock.Eq(expectedUser.Username)).
					Times(1).
					Return(expectedUser, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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
			buildStubs: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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
			buildStubs: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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
			buildStubs: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					GetUser(gomock.Any(), gomock.Eq(expectedUser.Username)).
					Times(1).
					Return(nil, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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
			buildStubs: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(expectedUser, nil)

				maker, ok := server.tokenMaker.(*mockdb.MockMaker)
				require.True(t, ok)

				maker.
					EXPECT().
					CreateToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return("", nil, errors.New("failed to create chacha20poly1305 cipher")).
					AnyTimes()
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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
			buildStubs: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(expectedUser, nil)

				maker, ok := server.tokenMaker.(*mockdb.MockMaker)
				require.True(t, ok)

				accessToken, accessPayload, err := createToken(
					server.config.TokenSymmetricKey,
					expectedUser.Username,
					expectedUser.Role,
					server.config.AccessTokenDuration,
				)
				maker.
					EXPECT().
					CreateToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(accessToken, accessPayload, err)

				maker.
					EXPECT().
					CreateToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(2).
					Return("", nil, errors.New("failed to create chacha20poly1305 cipher")).
					AnyTimes()
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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
			buildStubs: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					GetUser(gomock.Any(), gomock.Eq(expectedUser.Username)).
					Times(1).
					Return(expectedUser, nil).
					AnyTimes()

				store.
					EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrNoRows).
					AnyTimes()

				accessToken, accessPayload, err := createToken(
					server.config.TokenSymmetricKey,
					expectedUser.Username,
					expectedUser.Role,
					server.config.AccessTokenDuration,
				)

				maker, ok := server.tokenMaker.(*mockdb.MockMaker)
				require.True(t, ok)

				maker.
					EXPECT().
					CreateToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(accessToken, accessPayload, err).
					AnyTimes()

				refreshToken, refreshPayload, err := createToken(
					server.config.TokenSymmetricKey,
					expectedUser.Username,
					expectedUser.Role,
					server.config.RefreshTokenDuration,
				)
				maker.
					EXPECT().
					CreateToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(2).
					Return(refreshToken, refreshPayload, err).
					AnyTimes()
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			maker := mockdb.NewMockMaker(ctrl)

			server := newServer(store, maker, tc.env)
			tc.buildStubs(server)

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/users/login"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUserInfo(t *testing.T) {
	expectedUser := createRandomUser()

	testCases := []struct {
		name          string
		username      string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			name:     "OK",
			username: expectedUser.Username,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), expectedUser.Username).
					Times(1).
					Return(expectedUser, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					expectedUser.Username,
					expectedUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusNotFound",
			username: expectedUser.Username,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					expectedUser.Username,
					expectedUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "StatusInternalServerError",
			username: expectedUser.Username,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					expectedUser.Username,
					expectedUser.Role,
					time.Minute,
				)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(store)
			recorder := httptest.NewRecorder()

			url := "/api/auth/users/info"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func createRandomUser() *db.User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	if err != nil {
		return nil
	}

	return &db.User{
		Username:          util.RandomOwner(),
		HashedPassword:    hashedPassword,
		FullName:          fmt.Sprintf("%s %s", util.RandomOwner(), util.RandomOwner()),
		Email:             util.RandomEmail(),
		PasswordChangedAt: time.Now(),
		CreatedAt:         time.Now(),
		Role: &db.Role{
			InternalID:  uuid.New(),
			Name:        RoleAdmin,
			Description: "Administration's role",
			ExternalID:  "URE101",
			UpdatedAt:   time.Now(),
			CreatedAt:   time.Now(),
		},
	}
}

func createRandomUserAndPassword() (*db.User, string) {
	password := util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	if err != nil {
		return nil, ""
	}

	return &db.User{
		Username:          util.RandomOwner(),
		HashedPassword:    hashedPassword,
		FullName:          fmt.Sprintf("%s %s", util.RandomOwner(), util.RandomOwner()),
		Email:             util.RandomEmail(),
		PasswordChangedAt: time.Now(),
		CreatedAt:         time.Now(),
		Role: &db.Role{
			InternalID:  uuid.New(),
			Name:        RoleAdmin,
			Description: "Administration's role",
			ExternalID:  "URE101",
			UpdatedAt:   time.Now(),
			CreatedAt:   time.Now(),
		},
	}, password
}

func createToken(symmetricKey string, username string, role *db.Role, duration time.Duration) (string, *token.Payload, error) {
	tokenMaker, err := token.NewPasetoMaker(symmetricKey)
	if err != nil {
		return "", nil, err
	}
	return tokenMaker.CreateToken(username, role, duration)
}

func newTestServer(store db.Store) *Server {
	config, err := util.LoadConfig("../app.yaml", "test")
	if err != nil {
		log.Fatal().Msg("cannot load config")
	}
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil
	}

	return NewServer(store, tokenMaker, config)
}

func newServer(store db.Store, tokenMaker token.Maker, env string) *Server {
	config, err := util.LoadConfig("../app.yaml", env)
	if err != nil {
		log.Fatal().Msg("cannot load config")
	}

	return NewServer(store, tokenMaker, config)
}
