package api

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/newbri/posadamissportia/db"
	"github.com/newbri/posadamissportia/mocker"
	"github.com/newbri/posadamissportia/token"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthMiddleware(t *testing.T) {
	adminUser := createRandomUser(db.RoleAdmin, true)
	testCases := []struct {
		name     string
		username string
		env      string
		body     gin.H
		mock     func(server *Server)
		response func(recorder *httptest.ResponseRecorder)
		auth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			name:     "NoTokenNoAuthorization",
			env:      "test",
			username: adminUser.Username,
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				querier.
					On("CreateToken", mock.Anything, mock.Anything, mock.Anything).
					Return("", nil, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					"",
					adminUser.Username,
					adminUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "NoAuthorizationType",
			env:      "test",
			username: adminUser.Username,
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				refreshToken, _, err := createToken(
					server.config.GetConfig().TokenSymmetricKey,
					adminUser.Username,
					adminUser.Role,
					server.config.GetConfig().RefreshTokenDuration,
				)
				require.NoError(t, err)

				querier.
					On("CreateToken", mock.Anything, mock.Anything, mock.Anything).
					Return(refreshToken, nil, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					"",
					adminUser.Username,
					adminUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "NoAuthorizationTypeBearer",
			env:      "test",
			username: adminUser.Username,
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				refreshToken, _, err := createToken(
					server.config.GetConfig().TokenSymmetricKey,
					adminUser.Username,
					adminUser.Role,
					server.config.GetConfig().RefreshTokenDuration,
				)
				require.NoError(t, err)

				querier.
					On("CreateToken", mock.Anything, mock.Anything, mock.Anything).
					Return(refreshToken, nil, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					"bearer1",
					adminUser.Username,
					adminUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "NoToken",
			env:      "test",
			username: adminUser.Username,
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				refreshToken, payload, err := createToken(
					server.config.GetConfig().TokenSymmetricKey,
					adminUser.Username,
					adminUser.Role,
					server.config.GetConfig().RefreshTokenDuration,
				)
				require.NoError(t, err)

				querier.
					On("CreateToken", mock.Anything, mock.Anything, mock.Anything).
					Return(refreshToken, payload, nil)

				querier.
					On("VerifyToken", mock.Anything).
					Times(1).
					Return(nil, token.ErrInvalidToken)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					"",
					adminUser.Username,
					adminUser.Role,
					time.Minute,
				)
			},
		},
		{
			name:     "GetUser",
			env:      "test",
			username: adminUser.Username,
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				refreshToken, payload, err := createToken(
					server.config.GetConfig().TokenSymmetricKey,
					adminUser.Username,
					adminUser.Role,
					server.config.GetConfig().RefreshTokenDuration,
				)
				require.NoError(t, err)

				querier.
					On("CreateToken", mock.Anything, mock.Anything, mock.Anything).
					Times(1).
					Return(refreshToken, payload, nil)

				querier.
					On("VerifyToken", mock.Anything).
					Times(1).
					Return(payload, nil)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
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
			name:     "UserIsDeleted",
			env:      "test",
			username: adminUser.Username,
			mock: func(server *Server) {
				querier, ok := server.store.(*mocker.TestMocker)
				require.True(t, ok)

				refreshToken, payload, err := createToken(
					server.config.GetConfig().TokenSymmetricKey,
					adminUser.Username,
					adminUser.Role,
					server.config.GetConfig().RefreshTokenDuration,
				)
				require.NoError(t, err)

				querier.
					On("CreateToken", mock.Anything, mock.Anything, mock.Anything).
					Times(1).
					Return(refreshToken, payload, nil)

				querier.
					On("VerifyToken", mock.Anything).
					Times(1).
					Return(payload, nil)

				querier.
					On("GetUser", mock.Anything, mock.Anything).
					Times(1).
					Return(adminUser, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
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

			url := "/api/auth/customer/users/info"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.auth(t, request, server.tokenMaker)
			recorder := httptest.NewRecorder()
			server.router.ServeHTTP(recorder, request)
			tc.response(recorder)
		})
	}
}

//func TestPasetoAuthRole(t *testing.T) {
//	randomCustomer := createRandomUser(db.RoleCustomer, false)
//	testCases := []struct {
//		name     string
//		username string
//		env      string
//		body     gin.H
//		mock     func(server *Server)
//		response func(recorder *httptest.ResponseRecorder)
//		auth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
//	}{
//		{
//			name:     "NoAuthorizationPayloadKey",
//			env:      "test",
//			username: randomCustomer.Username,
//			mock: func(server *Server) {
//				querier, ok := server.store.(*mocker.TestMocker)
//				require.True(t, ok)
//
//				querier.
//					On("GetUser", mock.Anything, mock.Anything).
//					Return(randomCustomer, nil)
//
//				refreshToken, payload, err := createRandomToken(
//					randomCustomer.Username,
//					db.RoleCustomer,
//				)
//				require.NoError(t, err)
//
//				querier.
//					On("CreateToken", mock.Anything, mock.Anything, mock.Anything).
//					Return(refreshToken, payload, nil)
//
//				querier.
//					On("VerifyToken", mock.Anything).
//					Return(payload, nil)
//
//				config1 := createConfiguration()
//				config2 := createConfiguration()
//
//				querier.
//					On("GetConfig").
//					Times(1).
//					Return(config1)
//
//				config2.AuthorizationPayloadKey = "wrong"
//
//				querier.
//					On("GetConfig").
//					Times(2).
//					Return(config2)
//			},
//			response: func(recorder *httptest.ResponseRecorder) {
//				require.Equal(t, http.StatusUnauthorized, recorder.Code)
//			},
//			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
//				addAuthorization(t,
//					request,
//					tokenMaker,
//					authorizationTypeBearer,
//					randomCustomer.Username,
//					randomCustomer.Role,
//					time.Minute,
//				)
//			},
//		},
//		{
//			name:     "WrongPayload",
//			env:      "test",
//			username: randomCustomer.Username,
//			mock: func(server *Server) {
//				querier, ok := server.store.(*mocker.TestMocker)
//				require.True(t, ok)
//
//				querier.
//					On("GetUser", mock.Anything, mock.Anything).
//					Times(1).
//					Return(randomCustomer, nil)
//
//				refreshToken, payload, err := createRandomToken(
//					randomCustomer.Username,
//					db.RoleVisitor,
//				)
//				require.NoError(t, err)
//
//				querier.
//					On("CreateToken", mock.Anything, mock.Anything, mock.Anything).
//					Times(1).
//					Return(refreshToken, payload, nil)
//
//				querier.
//					On("VerifyToken", mock.Anything).
//					Times(1).
//					Return(payload, nil)
//
//				config1 := createConfiguration()
//				config2 := createConfiguration()
//
//				querier.
//					On("GetConfig").
//					Times(2).
//					Return(config1)
//
//				querier.
//					On("GetConfig").
//					Times(1).
//					Return(config2)
//			},
//			response: func(recorder *httptest.ResponseRecorder) {
//				require.Equal(t, http.StatusUnauthorized, recorder.Code)
//			},
//			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
//				addAuthorization(t,
//					request,
//					tokenMaker,
//					authorizationTypeBearer,
//					randomCustomer.Username,
//					randomCustomer.Role,
//					time.Minute,
//				)
//			},
//		},
//	}
//
//	for _, tc := range testCases {
//		t.Run(tc.name, func(t *testing.T) {
//			querier := new(mocker.TestMocker)
//			server := newTestServer(querier, tc.env)
//			tc.mock(server)
//
//			url := "/api/auth/customer/users/info"
//			request, err := http.NewRequest(http.MethodGet, url, nil)
//			require.NoError(t, err)
//
//			tc.auth(t, request, server.tokenMaker)
//			recorder := httptest.NewRecorder()
//			server.router.ServeHTTP(recorder, request)
//			tc.response(recorder)
//		})
//	}
//}
