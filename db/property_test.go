package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestCreateProperty(t *testing.T) {
	db, mocker, err := sqlmock.New()
	require.NoError(t, err)
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	arg := createPropertyParams()
	expectedProperty := createProperty(arg)
	testCases := []struct {
		name              string
		propertyQueryRows *sqlmock.Rows
		mock              func(userQueryRows *sqlmock.Rows, arg *CreatePropertyParams)
		response          func(querier Querier, expectedProperty *Property, arg CreatePropertyParams)
	}{
		{
			name:              "ValidPropertyCreation",
			propertyQueryRows: getMockedExpectedCreateProperty(expectedProperty),
			mock: func(userQueryRows *sqlmock.Rows, arg *CreatePropertyParams) {
				// the CreateUser sql mock
				mocker.
					ExpectQuery(regexp.QuoteMeta(createPropertyQuery)).
					WithArgs(
						arg.Name,
						arg.Address,
						arg.State,
						arg.City,
						arg.Country,
						arg.PostalCode,
						arg.Phone,
						arg.Email,
						arg.ExpiredAt,
						arg.CreatedAt,
					).
					WillReturnRows(userQueryRows)
			},
			response: func(querier Querier, expectedProperty *Property, arg CreatePropertyParams) {
				actualProperty, err := querier.CreateProperty(context.Background(), arg)
				require.NoError(t, err)
				require.Equal(t, actualProperty, expectedProperty)
			},
		},
		{
			name:              "InvalidPropertyCreation",
			propertyQueryRows: getMockedExpectedCreateProperty(expectedProperty),
			mock: func(userQueryRows *sqlmock.Rows, arg *CreatePropertyParams) {
				// the CreateUser sql mock
				mocker.
					ExpectQuery(regexp.QuoteMeta(createPropertyQuery)).
					WithArgs(
						arg.Name,
						arg.Address,
						arg.State,
						arg.City,
						arg.Country,
						arg.PostalCode,
						arg.Phone,
						arg.Email,
						arg.ExpiredAt,
						arg.CreatedAt,
					).WillReturnError(fmt.Errorf("sql: no rows in result set"))
			},
			response: func(querier Querier, expectedProperty *Property, arg CreatePropertyParams) {
				actualProperty, err := querier.CreateProperty(context.Background(), arg)
				require.Error(t, err)
				require.Nil(t, actualProperty)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockQuery := &Queries{db: db}

			tc.mock(tc.propertyQueryRows, arg)
			tc.response(mockQuery, expectedProperty, *arg)
		})
	}
}

func TestActivateDeactivateProperty(t *testing.T) {
	db, mocker, err := sqlmock.New()
	require.NoError(t, err)
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	expectedProperty := createProperty(createPropertyParams())
	testCases := []struct {
		name              string
		isActive          bool
		externalId        string
		propertyQueryRows *sqlmock.Rows
		mock              func(userQueryRows *sqlmock.Rows, isActive bool, externalId string)
		response          func(querier Querier, expectedProperty *Property, isActive bool, externalId string)
	}{
		{
			name:              "ActivateProperty",
			isActive:          true,
			externalId:        "PRO103",
			propertyQueryRows: getMockedExpectedProperty(expectedProperty),
			mock: func(propertyQueryRows *sqlmock.Rows, isActive bool, externalId string) {
				// the CreateUser sql mock
				mocker.
					ExpectQuery(regexp.QuoteMeta(activatePropertyQuery)).
					WithArgs(isActive, externalId).
					WillReturnRows(propertyQueryRows)
			},
			response: func(querier Querier, expectedProperty *Property, isActive bool, externalId string) {
				actualProperty, err := querier.ActivateDeactivateProperty(context.Background(), isActive, externalId)
				require.NoError(t, err)
				require.Equal(t, actualProperty, expectedProperty)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockQuery := &Queries{db: db}

			tc.mock(tc.propertyQueryRows, tc.isActive, tc.externalId)
			tc.response(mockQuery, expectedProperty, tc.isActive, tc.externalId)
		})
	}
}

func TestGetAllProperty(t *testing.T) {
	db, mocker, err := sqlmock.New()
	require.NoError(t, err)
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	expectedProperty := []*Property{createProperty(createPropertyParams())}
	testCases := []struct {
		name              string
		isActive          bool
		externalId        string
		arg               LimitOffset
		propertyQueryRows *sqlmock.Rows
		mock              func(userQueryRows *sqlmock.Rows, arg LimitOffset)
		response          func(querier Querier, expectedProperty []*Property, arg LimitOffset)
	}{
		{
			name:              "GetProperty",
			isActive:          true,
			externalId:        "PRO103",
			arg:               LimitOffset{Limit: 0, Offset: 0},
			propertyQueryRows: getMockedExpectedProperty(expectedProperty[0]),
			mock: func(propertyQueryRows *sqlmock.Rows, arg LimitOffset) {
				// the CreateUser sql mock
				mocker.
					ExpectQuery(regexp.QuoteMeta(getAllPropertyQuery)).
					WithArgs(arg.Limit, arg.Offset).
					WillReturnRows(propertyQueryRows)
			},
			response: func(querier Querier, expectedProperty []*Property, arg LimitOffset) {
				actualProperty, err := querier.GetAllProperty(context.Background(), arg)
				require.NoError(t, err)
				require.Equal(t, actualProperty, expectedProperty)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockQuery := &Queries{db: db}

			tc.mock(tc.propertyQueryRows, tc.arg)
			tc.response(mockQuery, expectedProperty, tc.arg)
		})
	}
}
