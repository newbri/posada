package api

import (
	"github.com/gin-gonic/gin"
	"github.com/newbri/posadamissportia/db"
	"github.com/newbri/posadamissportia/db/util"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{}
	server, err := NewServer(config, store)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
