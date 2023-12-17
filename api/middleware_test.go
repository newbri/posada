package api

import (
	"fmt"
	"github.com/newbri/posadamissportia/token"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func addAuthorization(t *testing.T, request *http.Request, tokenMaker token.Maker, authorizationType string, username string, duration time.Duration) {
	userToken, payload, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, userToken)
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
}
