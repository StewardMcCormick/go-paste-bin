package api_key

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

type Cache interface {
	Set(ctx context.Context, hash string, key *domain.APIKey)
	Get(ctx context.Context, hash string) *domain.APIKey
	DeleteByKey(ctx context.Context, hash string)
}

type Repository struct {
	Pool  postgres.DBTX
	Cache Cache
}

func NewRepository(pool postgres.DBTX, cache Cache) *Repository {
	return &Repository{pool, cache}
}

func (r *Repository) Create(ctx context.Context, userId int64, key *domain.APIKey) (*domain.APIKey, error) {
	query := `INSERT INTO api_key(key_hash, user_id, created_at, expire_at, key_prefix) 
					VALUES ($1, $2, $3, $4, $5) RETURNING expire_at`
	_, err := r.Pool.Exec(ctx, query, key.Key, userId, key.CreatedAt, key.ExpiresAt, key.Prefix)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, fmt.Errorf("api-key creating error - %w", errs.ErrAPIKeyAlreadyExists)
		}

		return nil, fmt.Errorf("api-key creating error - %w", err)
	}

	key.UserId = userId
	r.Cache.Set(ctx, key.Key, key)
	return key, nil
}

func (r *Repository) RevokeKeyByUserId(ctx context.Context, userId int64) error {
	query := `DELETE FROM api_key WHERE user_id=$1 RETURNING key_hash`

	hash := ""
	err := r.Pool.QueryRow(ctx, query, userId).Scan(&hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("api-key revoke error - %w", err)
	}

	r.Cache.DeleteByKey(ctx, hash)
	return nil
}

func (r *Repository) GetByKeyHash(ctx context.Context, hash string) (key *domain.APIKey, err error) {
	if keyFromCache := r.Cache.Get(ctx, hash); keyFromCache != nil {
		return keyFromCache, nil
	}

	query := `SELECT key_hash, user_id, created_at, expire_at, key_prefix FROM api_key WHERE key_hash=$1`

	key = &domain.APIKey{}
	err = r.Pool.QueryRow(ctx, query, hash).
		Scan(&key.Key, &key.UserId, &key.CreatedAt, &key.ExpiresAt, &key.Prefix)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("api-key get error - %w", errs.ErrAPIKeyNotFound)
		}
		return nil, fmt.Errorf("api-key get error - %w", err)
	}

	r.Cache.Set(ctx, key.Key, key)

	return key, nil
}
