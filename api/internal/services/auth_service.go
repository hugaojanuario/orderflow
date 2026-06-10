package services

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"orderflow/api/internal/models"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrEmailAlreadyExists = errors.New("email already exists")

// UserRepository é o contrato que o service precisa do repositório de usuários
type UserRepository interface {
	Create(user *models.User) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
}

// Claims customizadas do token de acesso
type Claims struct {
	UserId int `json:"user_id"`
	jwt.RegisteredClaims
}

type AuthService struct {
	repo      UserRepository
	jwtSecret string
}

func NewAuthService(repo UserRepository, jwtSecret string) *AuthService {
	return &AuthService{repo: repo, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(req *models.RegisterRequest) (*models.User, error) {
	existing, err := s.repo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar o usuário: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar hash da senha: %w", err)
	}

	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
	}

	created, err := s.repo.Create(user)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar o usuário: %w", err)
	}

	return created, nil
}

func (s *AuthService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.repo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar o usuário: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar o token: %w", err)
	}

	return &models.LoginResponse{Token: token, User: *user}, nil
}

func (s *AuthService) generateToken(user *models.User) (string, error) {
	claims := Claims{
		UserId: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(user.ID),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("erro ao assinar o token: %w", err)
	}

	return signed, nil
}
