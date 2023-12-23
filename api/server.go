package api

import (
	"github.com/gin-gonic/gin"
	"github.com/newbri/posadamissportia/db"
	"github.com/newbri/posadamissportia/db/util"
	"github.com/newbri/posadamissportia/token"
)

type Server struct {
	config     *util.Config
	store      db.Store
	router     *gin.Engine
	tokenMaker token.Maker
}

func NewServer(store db.Store, tokenMaker token.Maker, config *util.Config) *Server {
	server := &Server{store: store, tokenMaker: tokenMaker, config: config}

	server.setupRouter()
	return server
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)
	router.POST("/tokens/renew_access", server.renewAccessToken)

	router.POST("/role", server.createRole)
	router.GET("/role/:id", server.getRole)
	router.POST("/role/all", server.getAllRole)
	router.PUT("/role", server.updateRole)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))

	// add routes to router
	authRoutes.GET("/users/:username", server.getUser)
	authRoutes.GET("/users/info", server.getUserInfo)
	authRoutes.DELETE("/users/:username", server.deleteUser)
	authRoutes.PUT("/users", server.updateUser)

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
