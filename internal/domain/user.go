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
