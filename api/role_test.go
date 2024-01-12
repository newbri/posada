package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/newbri/posadamissportia/db"
	mockdb "github.com/newbri/posadamissportia/db/mock"
	"github.com/newbri/posadamissportia/token"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateRole(t *testing.T) {
	user := createRandomUser()
	expectedRole := &db.Role{
		InternalID:  uuid.New(),
		Name:        "visitor",
		Description: "Visitor's expectedRole",
		ExternalID:  "URE101",
		UpdatedAt:   time.Now(),
		CreatedAt:   time.Now(),
	}

	testCases := []struct {
		name         string
		env          string
		body         gin.H
		mock         func(server *Server)
		response     func(recorder *httptest.ResponseRecorder)
		authenticate func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			name: "OK",
			env:  "test",
			body: gin.H{
				"name":        expectedRole.Name,
				"description": expectedRole.Description,
			},
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					CreateRole(gomock.Any(), gomock.Any()).
					Times(1).
					Return(expectedRole, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)

				var request roleResponse
				err = json.Unmarshal(data, &request)
				require.NoError(t, err)

				require.Equal(t, expectedRole.Name, request.Name)
				require.Equal(t, expectedRole.Description, request.Description)
				require.Equal(t, expectedRole.ExternalID, request.ExternalID)
				require.Equal(t, expectedRole.UpdatedAt.Unix(), request.UpdatedAt.Unix())
				require.Equal(t, expectedRole.CreatedAt.Unix(), request.CreatedAt.Unix())
			},
			authenticate: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					user.Username,
					user.Role,
					time.Minute,
				)
			},
		},
		{
			name: "StatusBadRequest",
			env:  "test",
			body: gin.H{
				"name":         expectedRole.Name,
				"description1": expectedRole.Description,
			},
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					CreateRole(gomock.Any(), gomock.Any()).
					Times(0)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
			authenticate: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					user.Username,
					user.Role,
					time.Minute,
				)
			},
		},
		{
			name: "unique_violation",
			env:  "test",
			body: gin.H{
				"name":        expectedRole.Name,
				"description": expectedRole.Description,
			},
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					CreateRole(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, &pq.Error{Code: "23505"})
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
			authenticate: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					user.Username,
					user.Role,
					time.Minute,
				)
			},
		},
		{
			name: "InternalServerError",
			env:  "test",
			body: gin.H{
				"name":        expectedRole.Name,
				"description": expectedRole.Description,
			},
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					CreateRole(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			authenticate: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					user.Username,
					user.Role,
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
			server := newTestServer(store)
			tc.mock(server)

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/auth/admin/role"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			tc.authenticate(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.response(recorder)
		})
	}
}
