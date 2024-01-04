package api

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
)

var (
	ErrUniqueViolation   = errors.New("identifier must be unique")
	ErrInternalServer    = errors.New("an error occurs")
	ErrNoRow             = errors.New("no row was returned")
	ErrShouldBindUri     = errors.New("uri could not bind with the struct")
	ErrPasswordMistMach  = errors.New("password miss match")
	ErrVerifyToken       = errors.New("could not verify the token")
	ErrSession           = errors.New("session could not be created")
	ErrBlockedSession    = errors.New("blocked session")
	ErrWrongUserSession  = errors.New("incorrect user's session")
	ErrWrongSessionToken = errors.New("mismatched session token")
	ErrExpiredSession    = errors.New("session has expired")
	ErrTokenCreation     = errors.New("an issued occurs when creating token")
	ErrNoRole            = errors.New("role not found")
)

func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("the %s field is required", fe.Field())
	case "email":
		return fmt.Sprintf("the %s field is invalid", fe.Field())
	case "alphanum":
		return fmt.Sprintf("the %s field must be of type alphanumeric", fe.Field())
	case "min":
		return fmt.Sprintf("the %s field must has a minimum value of %s", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("the %s field value must be greater or equal to one", fe.Field())
	}
	return ""
}

func validateFieldError(ctx *gin.Context, err error) {
	type response struct {
		Field string
		Msg   string
	}
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make([]response, len(ve))
		for i, fe := range ve {
			out[i] = response{fe.Field(), msgForTag(fe)}
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"errors": out})
	}
}
