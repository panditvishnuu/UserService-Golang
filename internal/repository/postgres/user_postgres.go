package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/panditvishnuu/userservice/internal/domain"
	"github.com/panditvishnuu/userservice/internal/repository"
)

type userRepo struct {
	db *sql.DB
}

func New(db *sql.DB) repository.UserRepo {
	return &userRepo{db: db}
}

func NewWithPing(ctx context.Context, db *sql.DB) (repository.UserRepo, error) {
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	return &userRepo{db: db}, nil
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Name, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)

	var pqerr *pq.Error

	if err != nil {
		if errors.As(err, &pqerr) && pqerr.Code == "23505" {
			return &domain.EmailAlreadyExists{Email: user.Email}
		}
	}
	return fmt.Errorf("userRepo.Create: %w", err)
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
        SELECT id, name, email, password_hash, created_at, updated_at
        FROM users WHERE email = $1
    `
	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &domain.ErrorNotFound{Email: email}
		}
		return nil, fmt.Errorf("userRepository.GetByEmail: %w", err)
	}
	return user, nil
}

func (r *userRepo) GetByID(ctx context.Context, Id string) (*domain.User, error) {
	query := `
        SELECT id, name, email, password_hash, created_at, updated_at
        FROM users WHERE id = $1
    `
	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, Id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &domain.ErrorNotFound{UserID: Id}
		}
		return nil, fmt.Errorf("userRepo.GetByID: %w", err)
	}
	return user, nil
}
