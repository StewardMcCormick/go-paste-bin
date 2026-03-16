package api_key

import (
	"context"
	"fmt"

	"github.com/StewardMcCormick/Paste_Bin/internal/adapter/postgres"
	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
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
	log := appctx.GetLogger(ctx)

	query := `INSERT INTO api_key(key_hash, user_id, created_at, expire_at, key_prefix) 
					VALUES ($1, $2, $3, $4, $5) RETURNING expire_at`
	_, err := r.Pool.Exec(ctx, query, key.Key, userId, key.CreatedAt, key.ExpiresAt, key.Prefix)

	if err != nil {
		log.Error(err.Error())

		return nil, err
	}

	key.UserId = userId
	r.Cache.Set(ctx, key.Key, key)
	return key, nil
}

func (r *Repository) RevokeKeyByUserId(ctx context.Context, userId int64) error {
	log := appctx.GetLogger(ctx)

	query := `DELETE FROM api_key WHERE user_id=$1 RETURNING key_hash`

	hash := ""
	err := r.Pool.QueryRow(ctx, query, userId).Scan(&hash)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	r.Cache.DeleteByKey(ctx, hash)
	return nil
}

func (r *Repository) GetByKeyHash(ctx context.Context, hash string) (key *domain.APIKey, err error) {
	log := appctx.GetLogger(ctx)

	if keyFromCache := r.Cache.Get(ctx, hash); keyFromCache != nil {
		return keyFromCache, nil
	}

	query := `SELECT key_hash, user_id, created_at, expire_at, key_prefix FROM api_key WHERE key_hash=$1`
	rows, err := r.Pool.Query(ctx, query, hash)
	defer rows.Close()

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if !rows.Next() {
		return nil, nil
	}

	key = &domain.APIKey{}
	err = rows.Scan(&key.Key, &key.UserId, &key.CreatedAt, &key.ExpiresAt, &key.Prefix)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	r.Cache.Set(ctx, key.Key, key)

	log.Debug(fmt.Sprintf("Get key from db - %v", key))
	return key, nil
}
