package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/StewardMcCormick/Paste_Bin/internal/adapter/postgres"
	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repository struct {
	Pool postgres.DBTX
}

func NewRepository(pool postgres.DBTX) *Repository {
	return &Repository{Pool: pool}
}

func (r *Repository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	resultUser := &domain.User{}

	userQuery := `INSERT INTO users(username, password_hash, created_at) 
    				VALUES ($1, $2, $3) RETURNING *`
	err := r.Pool.QueryRow(ctx, userQuery, user.Username, user.Password, user.CreatedAt).
		Scan(&resultUser.Id, &resultUser.Username, &resultUser.Password, &resultUser.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, fmt.Errorf("api-key creating error - %w", errs.ErrUserAlreadyExists)
		}

		return nil, fmt.Errorf("api-key creating error - %w", err)
	}

	return resultUser, nil
}

func (r *Repository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `SELECT id, username, password_hash, created_at FROM users WHERE username = $1`

	resultUser := &domain.User{}
	err := r.Pool.QueryRow(ctx, query, username).
		Scan(&resultUser.Id, &resultUser.Username, &resultUser.Password, &resultUser.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user get error - %w", errs.ErrUserNotFound)
		}
		return nil, err
	}

	return resultUser, nil
}
