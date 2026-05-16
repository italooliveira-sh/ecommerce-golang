package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/italooliveira-sh/ecommerce-golang/internal/domain"
)

type UpdateAddressInput struct {
	Street  string
	City    string
	State   string
	ZipCode string
	Country string
}

type AddressRepository interface {
	CreateAddress(ctx context.Context, address *domain.Address) error
	ListAddressesByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Address, error)
	UpdateAddress(ctx context.Context, id uuid.UUID, input UpdateAddressInput) error
	DeleteAddress(ctx context.Context, id uuid.UUID) error
	SetDefaultAddress(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}
