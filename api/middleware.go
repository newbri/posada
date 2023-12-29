package api

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/newbri/posadamissportia/token"
	"net/http"
	"strings"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
	adminRole               = "admin"
)

var ErrInvalidAuthHeaderFormat = errors.New("invalid authorization header format")
var ErrAuthHeaderNotProvided = errors.New("authorization header is not provided")

// authMiddleware is a Gin middleware function that performs authentication based on a provided token.
func authMiddleware(tokenMarker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(ErrAuthHeaderNotProvided))
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(ErrInvalidAuthHeaderFormat))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		accessToken := fields[1]
		payload, err := tokenMarker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}

func pasetoAuthAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		data, exist := ctx.Get(authorizationPayloadKey)
		if !exist {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication data is required"})
			ctx.Abort()
			return
		}

		payload, ok := data.(*token.Payload)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication payload is required"})
			ctx.Abort()
			return
		}

		if payload.Role.Name != adminRole {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Only Administrator is allowed to perform this action"})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
