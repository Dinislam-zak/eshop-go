package domain

import "errors"

var (
	ErrNotFound      = errors.New("entity not found")
	ErrAlreadyExists = errors.New("entity already exists")
	ErrValidation    = errors.New("validation error")
	ErrInUse         = errors.New("entity is in use")
)