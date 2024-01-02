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
	router.Use(errorHandlingMiddleware())

	apiGroup := router.Group("/api")

	apiGroup.POST("/users", server.createUser)
	apiGroup.POST("/users/login", server.loginUser)
	apiGroup.POST("/tokens/renew_access", server.renewAccessToken)

	authGroup := apiGroup.Group("/auth")

	authGroup.Use(authMiddleware(server.tokenMaker))
	authGroup.GET("/users/info", server.getUserInfo)
	authGroup.PUT("/users", server.updateUser)

	// admin
	adminGroup := authGroup.Group("/admin")
	adminGroup.Use(pasetoAuthRole(RoleAdmin))
	adminGroup.POST("/role", server.createRole)
	adminGroup.GET("/role/:id", server.getRole)
	adminGroup.POST("/role/all", server.getAllRole)
	adminGroup.PUT("/role", server.updateRole)
	adminGroup.DELETE("/role/:id", server.deleteRole)
	adminGroup.GET("/users/:username", server.getUser)
	adminGroup.DELETE("/users/:username", server.deleteUser)

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
