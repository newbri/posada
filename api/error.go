package api

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
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
	ErrNoCustomerFound   = errors.New("no customer found")
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

type response struct {
	Field string `json:"field" binding:"omitempty"`
	Msg   string `json:"msg"`
}

func validateFieldError(err error) *[]response {
	var out []response
	var ve validator.ValidationErrors
	switch {
	case errors.As(err, &ve):
		for _, fe := range ve {
			out = append(out, response{fe.Field(), msgForTag(fe)})
		}
	default:
		out = append(out, response{Msg: err.Error()})
	}
	return &out
}
