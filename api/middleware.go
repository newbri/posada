package api

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/newbri/posadamissportia/token"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
)

var ErrInvalidAuthHeaderFormat = errors.New("invalid authorization header format")
var ErrAuthHeaderNotProvided = errors.New("authorization header is not provided")

// authMiddleware is a Gin middleware function that performs authentication based on a provided token.
func authMiddleware(server *Server) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := strings.TrimSpace(ctx.GetHeader(server.config.GetConfig().AuthorizationHeaderKey))
		if len(authorizationHeader) == 0 {
			log.Info().Msg(ctx.Error(ErrAuthHeaderNotProvided).Error())
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			log.Info().Msg(ctx.Error(ErrInvalidAuthHeaderFormat).Error())
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		accessToken := fields[1]
		payload, err := server.tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		user, err := server.store.GetUser(ctx, payload.Username)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		if user.IsDeleted {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(errors.New("token is invalid. User does not exists")))
			return
		}

		ctx.Set(server.config.GetConfig().AuthorizationPayloadKey, payload)
		ctx.Next()
	}
}

func pasetoAuthRole(server *Server, role string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		data, exist := ctx.Get(server.config.GetConfig().AuthorizationPayloadKey)
		if !exist {
			err := errors.New("authentication data is required")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		payload, _ := data.(*token.Payload)
		if payload.Role.Name != role {
			err := fmt.Errorf("only %s is allowed to perform this action", role)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, err)
			return
		}

		ctx.Next()
	}
}

func errorHandlingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		// only run if there are some errors to handle
		if len(ctx.Errors) > 0 {
			for _, err := range ctx.Errors {
				response := validateFieldError(err)
				switch {
				case errors.Is(err.Err, ErrUniqueViolation):
					ctx.JSON(http.StatusForbidden, gin.H{"errors": response})
				case errors.Is(err.Err, ErrInternalServer):
					ctx.JSON(http.StatusInternalServerError, gin.H{"errors": response})
				case errors.Is(err.Err, ErrNoRow):
					ctx.JSON(http.StatusNotFound, gin.H{"errors": response})
				case errors.Is(err.Err, ErrShouldBindUri):
					ctx.JSON(http.StatusBadRequest, gin.H{"errors": response})
				case errors.Is(err.Err, ErrVerifyToken) || errors.Is(err.Err, ErrAuthHeaderNotProvided) || errors.Is(err.Err, ErrInvalidAuthHeaderFormat):
					ctx.JSON(http.StatusUnauthorized, gin.H{"errors": response})
				case errors.Is(err.Err, ErrSession):
					ctx.JSON(http.StatusInternalServerError, gin.H{"errors": response})
				case errors.Is(err.Err, ErrBlockedSession):
					ctx.JSON(http.StatusUnauthorized, gin.H{"errors": response})
				case errors.Is(err.Err, ErrWrongUserSession):
					ctx.JSON(http.StatusUnauthorized, gin.H{"errors": response})
				case errors.Is(err.Err, ErrWrongSessionToken):
					ctx.JSON(http.StatusUnauthorized, gin.H{"errors": response})
				case errors.Is(err.Err, ErrExpiredSession):
					ctx.JSON(http.StatusUnauthorized, gin.H{"errors": response})
				case errors.Is(err.Err, ErrTokenCreation):
					ctx.JSON(http.StatusInternalServerError, gin.H{"errors": response})
				case errors.Is(err.Err, ErrPasswordMistMach):
					ctx.JSON(http.StatusUnauthorized, gin.H{"errors": response})
				case errors.Is(err.Err, ErrNoRole) || errors.Is(err.Err, ErrNoCustomerFound):
					ctx.JSON(http.StatusNotFound, gin.H{"errors": response})
				default:
					ctx.JSON(http.StatusBadRequest, gin.H{"errors": response})
				}
			}

			// once we handled all the errors, clear them from the gin context
			ctx.Errors.Last().Meta = nil
		}
	}
}

func CORSMiddleware(server *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", server.config.GetConfig().AccessControlAllowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Headers", server.config.GetConfig().AccessControlAllowHeaders)
		c.Writer.Header().Set("Access-Control-Allow-Methods", server.config.GetConfig().AccessControlAllowMethods)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
