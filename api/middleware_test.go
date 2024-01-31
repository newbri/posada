package api

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/newbri/posadamissportia/db"
	mockdb "github.com/newbri/posadamissportia/db/mock"
	"github.com/newbri/posadamissportia/token"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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
				maker, ok := server.tokenMaker.(*mockdb.MockMaker)
				require.True(t, ok)

				maker.
					EXPECT().
					CreateToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
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
				maker, ok := server.tokenMaker.(*mockdb.MockMaker)
				require.True(t, ok)

				refreshToken, _, err := createToken(
					server.config.GetConfig().TokenSymmetricKey,
					adminUser.Username,
					adminUser.Role,
					server.config.GetConfig().RefreshTokenDuration,
				)
				require.NoError(t, err)

				maker.
					EXPECT().
					CreateToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
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
				maker, ok := server.tokenMaker.(*mockdb.MockMaker)
				require.True(t, ok)

				refreshToken, _, err := createToken(
					server.config.GetConfig().TokenSymmetricKey,
					adminUser.Username,
					adminUser.Role,
					server.config.GetConfig().RefreshTokenDuration,
				)
				require.NoError(t, err)

				maker.
					EXPECT().
					CreateToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
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
				maker, ok := server.tokenMaker.(*mockdb.MockMaker)
				require.True(t, ok)

				refreshToken, payload, err := createToken(
					server.config.GetConfig().TokenSymmetricKey,
					adminUser.Username,
					adminUser.Role,
					server.config.GetConfig().RefreshTokenDuration,
				)
				require.NoError(t, err)

				maker.
					EXPECT().
					CreateToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(refreshToken, payload, nil)

				maker.
					EXPECT().
					VerifyToken(gomock.Any()).Times(1).
					Return(nil, token.ErrInvalidToken)
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
			name:     "GetUser",
			env:      "test",
			username: adminUser.Username,
			mock: func(server *Server) {
				maker, ok := server.tokenMaker.(*mockdb.MockMaker)
				require.True(t, ok)

				refreshToken, payload, err := createToken(
					server.config.GetConfig().TokenSymmetricKey,
					adminUser.Username,
					adminUser.Role,
					server.config.GetConfig().RefreshTokenDuration,
				)
				require.NoError(t, err)

				maker.
					EXPECT().
					CreateToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(refreshToken, payload, nil)

				maker.
					EXPECT().
					VerifyToken(gomock.Any()).Times(1).
					Return(payload, nil)

				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
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
				maker, ok := server.tokenMaker.(*mockdb.MockMaker)
				require.True(t, ok)

				refreshToken, payload, err := createToken(
					server.config.GetConfig().TokenSymmetricKey,
					adminUser.Username,
					adminUser.Role,
					server.config.GetConfig().RefreshTokenDuration,
				)
				require.NoError(t, err)

				maker.
					EXPECT().
					CreateToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(refreshToken, payload, nil)

				maker.
					EXPECT().
					VerifyToken(gomock.Any()).Times(1).
					Return(payload, nil)

				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			maker := mockdb.NewMockMaker(ctrl)
			server := newServer(store, maker, tc.env)
			tc.mock(server)

			url := "/api/auth/users/info"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.auth(t, request, server.tokenMaker)
			recorder := httptest.NewRecorder()
			server.router.ServeHTTP(recorder, request)
			tc.response(recorder)
		})
	}
}
