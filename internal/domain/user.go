package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	UserRoleCustomer UserRole = "customer"
	UserRoleAdmin    UserRole = "admin"
)

type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PasswordHash string
	Role         UserRole
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(name string, email string, passwordHash string) *User {
	return &User{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         UserRoleCustomer,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

type Address struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Street    string
	City      string
	State     string
	ZipCode   string
	Country   string
	IsDefault bool
}
