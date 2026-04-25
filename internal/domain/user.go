package domain

import (
	"fmt"
	"time"
)

var (
	ErrUserNotFound       = fmt.Errorf("user not found")
	ErrEmailAlreadyExists = fmt.Errorf("email already exists")
	ErrInvalidCredentials = fmt.Errorf("invalid credentials")
)

type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ErrorNotFound struct {
	UserID string
	Email  string
}

func (e *ErrorNotFound) Error() string {
	return fmt.Sprintf("user %s not found", e.UserID)
}

func (e *ErrorNotFound) Is(target error) bool {
	return target == ErrUserNotFound
}

type EmailAlreadyExists struct {
	Email string
}

func (e *EmailAlreadyExists) Error() string {
	return fmt.Sprintf("email %s already exists", e.Email)
}

func (e *EmailAlreadyExists) Is(target error) bool {
	return target == ErrEmailAlreadyExists
}

type InvalidCredentials struct {
}

func (e *InvalidCredentials) Error() string {
	return "invalid credentials"
}

func (e *InvalidCredentials) Is(target error) bool {
	return target == ErrInvalidCredentials
}
