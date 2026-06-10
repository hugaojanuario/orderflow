package services

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"orderflow/api/internal/models"
)

type mockUserRepository struct {
	users     map[string]*models.User
	createErr error
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{users: map[string]*models.User{}}
}

func (m *mockUserRepository) Create(user *models.User) (*models.User, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	created := *user
	created.ID = len(m.users) + 1
	m.users[user.Email] = &created
	return &created, nil
}

func (m *mockUserRepository) GetByEmail(email string) (*models.User, error) {
	user, ok := m.users[email]
	if !ok {
		return nil, nil
	}
	return user, nil
}

func TestRegisterCreatesUserWithHashedPassword(t *testing.T) {
	repo := newMockUserRepository()
	service := NewAuthService(repo, "test-secret")

	user, err := service.Register(&models.RegisterRequest{
		Name:     "Hugo",
		Email:    "hugo@test.local",
		Password: "senha123",
	})
	if err != nil {
		t.Fatalf("esperava sucesso, recebeu erro: %v", err)
	}
	if user.ID == 0 {
		t.Fatal("esperava id preenchido")
	}
	if user.PasswordHash == "senha123" {
		t.Fatal("senha não pode ser salva em texto puro")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("senha123")); err != nil {
		t.Fatalf("hash da senha inválido: %v", err)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	repo := newMockUserRepository()
	service := NewAuthService(repo, "test-secret")

	req := &models.RegisterRequest{Name: "Hugo", Email: "hugo@test.local", Password: "senha123"}
	if _, err := service.Register(req); err != nil {
		t.Fatalf("primeiro registro deveria funcionar: %v", err)
	}

	_, err := service.Register(req)
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Fatalf("esperava ErrEmailAlreadyExists, recebeu: %v", err)
	}
}

func TestLoginSuccess(t *testing.T) {
	repo := newMockUserRepository()
	service := NewAuthService(repo, "test-secret")

	if _, err := service.Register(&models.RegisterRequest{Name: "Hugo", Email: "hugo@test.local", Password: "senha123"}); err != nil {
		t.Fatalf("erro ao registrar: %v", err)
	}

	response, err := service.Login(&models.LoginRequest{Email: "hugo@test.local", Password: "senha123"})
	if err != nil {
		t.Fatalf("esperava sucesso, recebeu erro: %v", err)
	}
	if response.Token == "" {
		t.Fatal("esperava token preenchido")
	}
	if response.User.Email != "hugo@test.local" {
		t.Fatalf("esperava email do usuário, recebeu: %s", response.User.Email)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	repo := newMockUserRepository()
	service := NewAuthService(repo, "test-secret")

	if _, err := service.Register(&models.RegisterRequest{Name: "Hugo", Email: "hugo@test.local", Password: "senha123"}); err != nil {
		t.Fatalf("erro ao registrar: %v", err)
	}

	_, err := service.Login(&models.LoginRequest{Email: "hugo@test.local", Password: "senha-errada"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("esperava ErrInvalidCredentials, recebeu: %v", err)
	}

	_, err = service.Login(&models.LoginRequest{Email: "nao-existe@test.local", Password: "senha123"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("esperava ErrInvalidCredentials para usuário inexistente, recebeu: %v", err)
	}
}
