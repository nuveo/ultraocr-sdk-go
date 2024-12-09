// Package common implements constants and errors.
package common

import "errors"

// SDK Errors.
var (
	ErrMountingRequest    = errors.New("failed to mount request")
	ErrDoingRequest       = errors.New("failed to request")
	ErrInvalidStatusCode  = errors.New("invalid status code")
	ErrParsingRequestBody = errors.New("failed to parse request body")
	ErrParsingResponse    = errors.New("failed to parse response body")
	ErrReadFile           = errors.New("failed to read file")
)
