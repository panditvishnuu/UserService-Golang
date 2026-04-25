package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/panditvishnuu/userservice/internal/domain"
	"github.com/panditvishnuu/userservice/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, name, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	GetByID(ctx context.Context, Id string) (*domain.User, error)
}

type userService struct {
	repo      repository.UserRepo
	jwtSecret string
	jwtExpiry time.Duration
	dummyHash []byte
}

func NewUserService(repo repository.UserRepo, jwtSecret string, jwtExpiryHours int) (UserService, error) {
	dummy, err := bcrypt.GenerateFromPassword([]byte(uuid.New().String()), 12)
	if err != nil {
		return nil, fmt.Errorf("NewUserService: generating dummy hash: %w", err)
	}

	return &userService{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpiry: time.Duration(jwtExpiryHours) * time.Hour,
		dummyHash: dummy,
	}, nil
}

func (s *userService) Register(ctx context.Context, name, email, password string) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("service.Register: hashing password: %w", err)
	}
	now := time.Now().UTC()
	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("service.Register: %w", err)
	}
	return user, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, email)

	var hashToCompare []byte

	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			hashToCompare = s.dummyHash
		} else {
			return "", fmt.Errorf("service.Login: %w", err)
		}
	} else {
		hashToCompare = []byte(user.PasswordHash)
	}

	// Always perform bcrypt comparison
	if bcrypt.CompareHashAndPassword(hashToCompare, []byte(password)) != nil {
		return "", &domain.InvalidCredentials{}
	}

	// If user didn't exist, still fail
	if err != nil {
		return "", &domain.InvalidCredentials{}
	}

	token, err := s.generateJWT(user)
	if err != nil {
		return "", fmt.Errorf("service.Login: %w", err)
	}

	return token, nil
}

func (s *userService) GetByID(ctx context.Context, Id string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, Id)
	if err != nil {
		return nil, fmt.Errorf("service.GetByID: %w", err)
	}
	return user, nil
}

// generate JWT
type claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func (s *userService) generateJWT(user *domain.User) (string, error) {
	now := time.Now()

	c := &claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("generateJWT: signing token: %w", err)
	}
	return signed, nil
}
