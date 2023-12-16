package api

import (
	"github.com/gin-gonic/gin"
	"github.com/newbri/posadamissportia/db"
	"github.com/newbri/posadamissportia/db/util"
	"github.com/newbri/posadamissportia/token"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	config, err := util.LoadConfig("../app.yaml", "test")
	if err != nil {
		log.Fatal().Msg("cannot load config")
	}
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil
	}

	server, err := NewServer(store, tokenMaker, config)
	require.NoError(t, err)

	return server
}

func newServer(t *testing.T, store db.Store, tokenMaker token.Maker, env string) *Server {
	config, err := util.LoadConfig("../app.yaml", env)
	if err != nil {
		log.Fatal().Msg("cannot load config")
	}
	server, err := NewServer(store, tokenMaker, config)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
