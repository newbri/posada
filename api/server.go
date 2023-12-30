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

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker)).Use(pasetoAuthCustomer())

	// add routes to router
	authRoutes.GET("/users/:username", server.getUser)
	authRoutes.GET("/users/info", server.getUserInfo)
	authRoutes.PUT("/users", server.updateUser)

	// admin
	adminRoutes := router.Group("/admin")
	adminRoutes.Use(authMiddleware(server.tokenMaker)).Use(pasetoAuthAdmin())
	adminRoutes.POST("/role", server.createRole)
	adminRoutes.GET("/role/:id", server.getRole)
	adminRoutes.POST("/role/all", server.getAllRole)
	adminRoutes.PUT("/role", server.updateRole)
	adminRoutes.DELETE("/role/:id", server.deleteRole)
	adminRoutes.DELETE("/users/:username", server.deleteUser)

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
