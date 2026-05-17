package service

import (
	"context"
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/italooliveira-sh/ecommerce-golang/internal/domain"
	"github.com/italooliveira-sh/ecommerce-golang/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyExists = errors.New("Email já cadastrado")
)

type UserService struct {
	repo     repository.UserRepository
	validate validator.Validate
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{
		repo:     repo,
		validate: *validator.New(),
	}
}

type CreateUserRequest struct {
	Name     string `validate:"required"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=6"`
}

type UserResponse struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Role      domain.UserRole
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *UserService) Register(ctx context.Context, req CreateUserRequest) (*UserResponse, error) {
	err := s.validate.Struct(req)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrEmailAlreadyExists
	}

	if err == repository.ErrNotFound {
		newUser, err := buildUser(req)
		if err != nil {
			return nil, err
		}

		user, err := s.repo.CreateUser(ctx, newUser)
		if err != nil {
			return nil, err
		}
		return &UserResponse{
			user.ID,
			user.Name,
			user.Email,
			user.Role,
			user.CreatedAt,
			user.UpdatedAt,
		}, nil
	}
	return nil, err
}

func buildUser(req CreateUserRequest) (*domain.User, error) {
	hash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	user := domain.NewUser(req.Name, req.Email, hash)
	return user, nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
