package api

import "errors"

var (
	ErrBindToJSON        = errors.New("type cannot be bind to json")
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
)
