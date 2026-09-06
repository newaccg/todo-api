package errors

import "errors"

var (
	ErrUserNotFound = errors.New("user with given email and password not found")
	ErrTaskNotFound = errors.New("task with given ID not found")

	ErrInvalidID         = errors.New("invalid ID")
	ErrInvalidJSON       = errors.New("invalid JSON")
	ErrInvalidToken      = errors.New("invalid JWT")
	ErrInvalidSignMethod = errors.New("invalid signing method")
	ErrInvalidURLValue   = errors.New("invalid URL value(s)")
	ErrInvalidOrder      = errors.New("invalid selected order")

	ErrTokenExpired = errors.New("JWT expired")

	ErrEmailExists = errors.New("given email is already in database")

	ErrEmptyField = errors.New("empty field in given input")
	ErrEmptyToken = errors.New("got empty token")

	ErrWrongPassword = errors.New("password from DB and given password don't match")

	ErrTooManyRequests = errors.New("too many requests")
)
