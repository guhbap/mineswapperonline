package utils

import "errors"

var (
	ErrRoomNameRequired  = errors.New("room name required")
	ErrInvalidDimensions = errors.New("rows and cols must be between 5 and 50")
	ErrInvalidMinesCount = errors.New("mines must be between 1 and (rows*cols-1)")
	ErrAuthRequired      = errors.New("username and password are required")
	ErrPasswordTooShort  = errors.New("password must be at least 6 characters")
)
