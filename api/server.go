package api

import (
	"github.com/gin-gonic/gin"
	"github.com/newbri/posadamissportia/configuration"
	"github.com/newbri/posadamissportia/db"
	"github.com/newbri/posadamissportia/token"
)

type Server struct {
	config     configuration.Configuration
	store      db.Store
	router     *gin.Engine
	tokenMaker token.Maker
}

func NewServer(store db.Store, tokenMaker token.Maker, config configuration.Configuration) *Server {
	server := &Server{store: store, tokenMaker: tokenMaker, config: config}

	server.setupRouter()
	return server
}

func (server *Server) setupRouter() {
	router := gin.Default()
	router.Use(CORSMiddleware(server), errorHandlingMiddleware())

	apiGroup := router.Group("/api")
	apiVersion := apiGroup.Group("/v1")

	apiVersion.POST("/users", server.createUser)
	apiVersion.POST("/users/login", server.loginUser)
	apiVersion.POST("/tokens/renew_access", server.renewAccessToken)

	authGroup := apiVersion.Group("/auth")
	authGroup.Use(authMiddleware(server))

	customerGroup := authGroup.Group("/customer")
	customerGroup.Use(pasetoAuthRole(server, db.RoleCustomer))
	customerGroup.GET("/users/info", server.getUserInfo)
	customerGroup.PUT("/users", server.updateUser)
	/**
	GET /users
	GET /users/:user
	POST /users
	POST/PUT /users/:user
	DELETE /users/:user
	*/

	// admin
	adminGroup := authGroup.Group("/admin")
	adminGroup.Use(pasetoAuthRole(server, db.RoleAdmin))
	adminGroup.POST("/role", server.createRole)
	adminGroup.GET("/role/:id", server.getRole)
	adminGroup.POST("/role/all", server.getAllRole)
	adminGroup.PUT("/role", server.updateRole)
	adminGroup.DELETE("/role/:id", server.deleteRole)
	adminGroup.GET("/users/:username", server.getUser)
	adminGroup.DELETE("/users/:username", server.deleteUser)
	adminGroup.GET("/users/info", server.getUserInfo)
	adminGroup.PUT("/users", server.updateUser)
	adminGroup.POST("/users/all/customer", server.getAllCustomer)
	adminGroup.POST("/property", server.createProperty)
	adminGroup.POST("/property/activate", server.activateDeactivateProperty)
	adminGroup.GET("/property/all", server.getAllProperty)

	// su
	suGroup := authGroup.Group("/su")
	suGroup.Use(pasetoAuthRole(server, db.RoleSuperUser))
	suGroup.POST("/role", server.createRole)
	suGroup.GET("/role/:id", server.getRole)
	suGroup.POST("/role/all", server.getAllRole)
	suGroup.PUT("/role", server.updateRole)
	suGroup.DELETE("/role/:id", server.deleteRole)
	suGroup.GET("/users/:username", server.getUser)
	suGroup.DELETE("/users/:username", server.deleteUser)
	suGroup.POST("/users/all/customer", server.getAllCustomer)
	suGroup.POST("/users/all/admin", server.getAllAdmin)
	suGroup.GET("/users/info", server.getUserInfo)
	suGroup.PUT("/users", server.updateUser)

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
