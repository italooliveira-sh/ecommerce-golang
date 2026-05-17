package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/italooliveira-sh/ecommerce-golang/internal/domain"
	"github.com/italooliveira-sh/ecommerce-golang/internal/repository"
)

type mockRepo struct {
	getUserByEmail func(email string) (domain.User, error)
	createUser     func(user *domain.User) (domain.User, error)
}

func (r *mockRepo) CreateUser(ctx context.Context, user *domain.User) (domain.User, error) {
	return r.createUser(user)
}

func (r *mockRepo) GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	return domain.User{}, nil
}

func (r *mockRepo) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	return r.getUserByEmail(email)
}

func (r *mockRepo) UpdateUserName(ctx context.Context, id uuid.UUID, name string) error {
	return nil
}

func (r *mockRepo) UpdateUserPassword(ctx context.Context, id uuid.UUID, password string) error {
	return nil
}

func TestCreateUserShouldReturnErrorWhenEmailExists(t *testing.T) {
	repo := &mockRepo{
		getUserByEmail: func(email string) (domain.User, error) {
			return domain.User{
				Email: "email@email.com",
			}, nil
		},
	}

	svc := NewUserService(repo)

	req := CreateUserRequest{
		Name:     "Fabricio Paiva",
		Email:    "email@email.com",
		Password: "1234",
	}

	_, err := svc.Register(context.Background(), req)

	if err == nil {
		t.Error("Esperava error, mas não veio nenhum")
	}
}

func TestCreateUserShouldHashPassword(t *testing.T) {
	var passwordHash string
	repo := &mockRepo{
		createUser: func(user *domain.User) (domain.User, error) {
			passwordHash = user.PasswordHash
			return domain.User{
				PasswordHash: passwordHash,
			}, nil
		},
		getUserByEmail: func(email string) (domain.User, error) {
			return domain.User{}, repository.ErrNotFound
		},
	}

	svc := NewUserService(repo)

	req := CreateUserRequest{
		Name:     "Fabricio Paiva",
		Email:    "email@email.com",
		Password: "123456",
	}

	_, err := svc.Register(context.Background(), req)

	if err != nil {
		t.Error("Erro inesperado")
	}

	if req.Password == passwordHash {
		t.Error("Senha não foi cryptografada!")
	}

}

func TestCreateUserShouldReturnErrorWhenNameIsEmpty(t *testing.T) {
	repo := &mockRepo{}
	svc := NewUserService(repo)

	req := CreateUserRequest{
		Name:     "",
		Email:    "email@email.com",
		Password: "123456",
	}

	_, err := svc.Register(context.Background(), req)
	if err == nil {
		t.Error("Esperava error, mas não veio nenhum")
	}
}

func TestCreateUserShouldReturnErrorWhenPasswordIsTooShort(t *testing.T) {
	repo := &mockRepo{}
	svc := NewUserService(repo)

	req := CreateUserRequest{
		Name:     "Italo",
		Email:    "email@email.com",
		Password: "12345",
	}

	_, err := svc.Register(context.Background(), req)
	if err == nil {
		t.Error("Esperava error, mas não veio nenhum")
	}
}

func TestCreateUserShouldReturnErrorWhenEmailIsInvalid(t *testing.T) {
	repo := &mockRepo{}
	svc := NewUserService(repo)

	req := CreateUserRequest{
		Name:     "Italo",
		Email:    "emailInvalido",
		Password: "123456",
	}

	_, err := svc.Register(context.Background(), req)
	if err == nil {
		t.Error("Esperava error, mas não veio nenhum")
	}
}
