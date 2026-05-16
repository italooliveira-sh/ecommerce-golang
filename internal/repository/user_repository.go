package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/italooliveira-sh/ecommerce-golang/internal/domain"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) (domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	UpdateUserName(ctx context.Context, id uuid.UUID, name string) error
	UpdateUserPassword(ctx context.Context, id uuid.UUID, password string) error
}
