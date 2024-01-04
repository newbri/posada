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
	RoleAdmin               = "admin"
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

func pasetoAuthRole(role string) gin.HandlerFunc {
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

		if payload.Role.Name != role {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Only %s is allowed to perform this action", role)})
			ctx.Abort()
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
				switch {
				case errors.Is(err.Err, ErrUniqueViolation):
					ctx.JSON(http.StatusForbidden, errorResponse(ErrUniqueViolation))
				case errors.Is(err.Err, ErrInternalServer):
					ctx.JSON(http.StatusInternalServerError, errorResponse(ErrInternalServer))
				case errors.Is(err.Err, ErrNoRow):
					ctx.JSON(http.StatusNotFound, errorResponse(ErrNoRow))
				case errors.Is(err.Err, ErrShouldBindUri):
					ctx.JSON(http.StatusBadRequest, errorResponse(ErrShouldBindUri))
				case errors.Is(err.Err, ErrVerifyToken):
					ctx.JSON(http.StatusUnauthorized, errorResponse(ErrVerifyToken))
				case errors.Is(err.Err, ErrSession):
					ctx.JSON(http.StatusInternalServerError, errorResponse(ErrSession))
				case errors.Is(err.Err, ErrBlockedSession):
					ctx.JSON(http.StatusUnauthorized, errorResponse(ErrBlockedSession))
				case errors.Is(err.Err, ErrWrongUserSession):
					ctx.JSON(http.StatusUnauthorized, errorResponse(ErrWrongUserSession))
				case errors.Is(err.Err, ErrWrongSessionToken):
					ctx.JSON(http.StatusUnauthorized, errorResponse(ErrWrongSessionToken))
				case errors.Is(err.Err, ErrExpiredSession):
					ctx.JSON(http.StatusUnauthorized, errorResponse(ErrExpiredSession))
				case errors.Is(err.Err, ErrTokenCreation):
					ctx.JSON(http.StatusInternalServerError, errorResponse(ErrTokenCreation))
				case errors.Is(err.Err, ErrPasswordMistMach):
					ctx.JSON(http.StatusUnauthorized, errorResponse(ErrPasswordMistMach))
				case errors.Is(err.Err, ErrNoRole):
					ctx.JSON(http.StatusNotFound, errorResponse(ErrNoRole))
				default:
					validateFieldError(ctx, err)
				}
			}

			// once we handled all the errors, clear them from the gin context
			ctx.Errors.Last().Meta = nil
		}
	}
}
