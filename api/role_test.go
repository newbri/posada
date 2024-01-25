package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/newbri/posadamissportia/db"
	mockdb "github.com/newbri/posadamissportia/db/mock"
	"github.com/newbri/posadamissportia/token"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateRole(t *testing.T) {
	user := createRandomUser(db.RoleAdmin)
	expectedRole := createRandomRole(db.RoleVisitor)

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
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
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
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
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
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
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
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
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
			tc.auth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.response(recorder)
		})
	}
}

func TestGetAllRole(t *testing.T) {
	roles := testGetAllRole()
	user := createRandomUser(db.RoleAdmin)

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
				"offset": 2,
				"limit":  2,
			},
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					GetAllRole(gomock.Any(), gomock.Any()).
					Times(1).
					Return(roles, nil).
					AnyTimes()
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)

				var request []roleResponse
				err = json.Unmarshal(data, &request)
				require.NoError(t, err)

				for i := range request {
					require.Equal(t, roles[i].ExternalID, request[i].ExternalID)
					require.Equal(t, roles[i].Name, request[i].Name)
					require.Equal(t, roles[i].Description, request[i].Description)
					require.Equal(t, roles[i].CreatedAt.Unix(), request[i].CreatedAt.Unix())
					require.Equal(t, roles[i].CreatedAt.Unix(), request[i].UpdatedAt.Unix())
				}
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
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
			name: "BadRequest",
			env:  "test",
			body: gin.H{
				"offset": 2,
				"limit1": 2,
			},
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					GetAllRole(gomock.Any(), gomock.Any()).
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
					user.Username,
					user.Role,
					time.Minute,
				)
			},
		},
		{
			name: "BadRequest",
			env:  "test",
			body: gin.H{
				"offset": 2,
				"limit":  2,
			},
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					GetAllRole(gomock.Any(), gomock.Any()).
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
					user.Username,
					user.Role,
					time.Minute,
				)
			},
		},
		{
			name: "ErrInternalServer",
			env:  "test",
			body: gin.H{
				"offset": 2,
				"limit":  2,
			},
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					GetAllRole(gomock.Any(), gomock.Any()).
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

			url := "/api/auth/admin/role/all"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.auth(t, request, server.tokenMaker)
			recorder := httptest.NewRecorder()
			server.router.ServeHTTP(recorder, request)
			tc.response(recorder)
		})
	}
}

func TestGetRole(t *testing.T) {
	role := testGetAllRole()[0]
	user := createRandomUser(db.RoleAdmin)
	testCases := []struct {
		name       string
		externalID string
		env        string
		mock       func(server *Server)
		response   func(recorder *httptest.ResponseRecorder)
		auth       func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			name:       "OK",
			externalID: role.ExternalID,
			env:        "test",
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					GetRole(gomock.Any(), gomock.Any()).
					Times(1).
					Return(role, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)

				var request roleResponse
				err = json.Unmarshal(data, &request)
				require.NoError(t, err)

				require.Equal(t, role.ExternalID, request.ExternalID)
				require.Equal(t, role.Name, request.Name)
				require.Equal(t, role.Description, request.Description)
				require.Equal(t, role.CreatedAt.Unix(), request.CreatedAt.Unix())
				require.Equal(t, role.CreatedAt.Unix(), request.UpdatedAt.Unix())
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
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
			name:       "BadRequest",
			externalID: "-@anewball",
			env:        "test",
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					GetRole(gomock.Any(), gomock.Any()).
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
					user.Username,
					user.Role,
					time.Minute,
				)
			},
		},
		{
			name:       "StatusNotFound",
			externalID: role.ExternalID,
			env:        "test",
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					GetRole(gomock.Any(), gomock.Any()).
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
					user.Username,
					user.Role,
					time.Minute,
				)
			},
		},
		{
			name:       "InternalServerError",
			externalID: role.ExternalID,
			env:        "test",
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					GetRole(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/auth/admin/role/%s", tc.externalID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.auth(t, request, server.tokenMaker)
			recorder := httptest.NewRecorder()
			server.router.ServeHTTP(recorder, request)
			tc.response(recorder)
		})
	}
}

func TestUpdateRole(t *testing.T) {
	role := testGetAllRole()[0]
	user := createRandomUser(db.RoleAdmin)
	testCases := []struct {
		name     string
		env      string
		body     gin.H
		mock     func(server *Server)
		response func(recorder *httptest.ResponseRecorder)
		auth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			name: "NameOK",
			env:  "test",
			body: gin.H{
				"external_id": role.ExternalID,
				"name":        role.Name,
			},
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					UpdateRole(gomock.Any(), gomock.Any()).
					Times(1).
					Return(role, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)

				var request roleResponse
				err = json.Unmarshal(data, &request)
				require.NoError(t, err)

				require.Equal(t, role.ExternalID, request.ExternalID)
				require.Equal(t, role.Name, request.Name)
				require.Equal(t, role.Description, request.Description)
				require.Equal(t, role.CreatedAt.Unix(), request.CreatedAt.Unix())
				require.Equal(t, role.UpdatedAt.Unix(), request.UpdatedAt.Unix())
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
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
			name: "DescriptionOK",
			env:  "test",
			body: gin.H{
				"external_id": role.ExternalID,
				"description": role.Description,
			},
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					UpdateRole(gomock.Any(), gomock.Any()).
					Times(1).
					Return(role, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)

				var request roleResponse
				err = json.Unmarshal(data, &request)
				require.NoError(t, err)

				require.Equal(t, role.ExternalID, request.ExternalID)
				require.Equal(t, role.Name, request.Name)
				require.Equal(t, role.Description, request.Description)
				require.Equal(t, role.CreatedAt.Unix(), request.CreatedAt.Unix())
				require.Equal(t, role.CreatedAt.Unix(), request.UpdatedAt.Unix())
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
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
			name: "BadRequest",
			env:  "test",
			body: gin.H{
				"external_id1": role.ExternalID,
				"description":  role.Description,
			},
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					UpdateRole(gomock.Any(), gomock.Any()).
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
					user.Username,
					user.Role,
					time.Minute,
				)
			},
		},
		{
			name: "StatusNotFound",
			env:  "test",
			body: gin.H{
				"external_id": role.ExternalID,
				"description": role.Description,
			},
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					UpdateRole(gomock.Any(), gomock.Any()).
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
					user.Username,
					user.Role,
					time.Minute,
				)
			},
		},
		{
			name: "StatusInternalServer",
			env:  "test",
			body: gin.H{
				"external_id": role.ExternalID,
				"description": role.Description,
			},
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					UpdateRole(gomock.Any(), gomock.Any()).
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
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.auth(t, request, server.tokenMaker)
			recorder := httptest.NewRecorder()
			server.router.ServeHTTP(recorder, request)
			tc.response(recorder)
		})
	}
}

func TestDeleteRole(t *testing.T) {
	role := testGetAllRole()[0]
	user := createRandomUser(db.RoleAdmin)
	testCases := []struct {
		name       string
		externalID string
		env        string
		mock       func(server *Server)
		response   func(recorder *httptest.ResponseRecorder)
		auth       func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			name:       "OK",
			externalID: role.ExternalID,
			env:        "test",
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					DeleteRole(gomock.Any(), gomock.Any()).
					Times(1).
					Return(role, nil)
			},
			response: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)

				var request roleResponse
				err = json.Unmarshal(data, &request)
				require.NoError(t, err)

				require.Equal(t, role.ExternalID, request.ExternalID)
				require.Equal(t, role.Name, request.Name)
				require.Equal(t, role.Description, request.Description)
				require.Equal(t, role.CreatedAt.Unix(), request.CreatedAt.Unix())
				require.Equal(t, role.UpdatedAt.Unix(), request.UpdatedAt.Unix())
			},
			auth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
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
			name:       "BadRequest",
			externalID: "-@anewball",
			env:        "test",
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					DeleteRole(gomock.Any(), gomock.Any()).
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
					user.Username,
					user.Role,
					time.Minute,
				)
			},
		},
		{
			name:       "StatusNotFound",
			externalID: role.ExternalID,
			env:        "test",
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					DeleteRole(gomock.Any(), gomock.Any()).
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
					user.Username,
					user.Role,
					time.Minute,
				)
			},
		},
		{
			name:       "StatusInternalServerError",
			externalID: role.ExternalID,
			env:        "test",
			mock: func(server *Server) {
				store, ok := server.store.(*mockdb.MockStore)
				require.True(t, ok)

				store.
					EXPECT().
					DeleteRole(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/auth/admin/role/%s", tc.externalID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.auth(t, request, server.tokenMaker)
			recorder := httptest.NewRecorder()
			server.router.ServeHTTP(recorder, request)
			tc.response(recorder)
		})
	}
}

func testGetAllRole() []*db.Role {
	var roles []*db.Role
	roles = append(roles, createRandomRole(db.RoleAdmin))
	roles = append(roles, createRandomRole(db.RoleVisitor))
	roles = append(roles, createRandomRole(db.RoleCustomer))
	return roles
}
