package api

import "errors"

var (
	ErrBindToJSON       = errors.New("type cannot be bind to json")
	ErrUniqueViolation  = errors.New("identifier must be unique")
	ErrInternalServer   = errors.New("an error occurs")
	ErrNoRow            = errors.New("no row was returned")
	ErrShouldBindUri    = errors.New("uri could not bind with the struct")
	ErrPasswordMistMach = errors.New("password miss match")
)
