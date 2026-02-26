package api_key

import (
	"context"
	"fmt"

	"github.com/StewardMcCormick/Paste_Bin/internal/adapter/postgres"
	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
)

type Repository struct {
	Pool postgres.DBTX
}

func (r *Repository) Create(ctx context.Context, userId int64, key *domain.APIKey) (*domain.APIKey, error) {
	log := appctx.GetLogger(ctx)

	query := `INSERT INTO api_key(key_hash, user_id, created_at, expire_at, key_prefix) 
					VALUES ($1, $2, $3, $4, $5) RETURNING (expire_at)`
	_, err := r.Pool.Exec(ctx, query, key.Key, userId, key.CreatedAt, key.ExpiresAt, key.Prefix)

	if err != nil {
		log.Error(err.Error())

		return nil, err
	}

	return key, nil
}

func (r *Repository) GetByHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	log := appctx.GetLogger(ctx)

	query := `SELECT (key_hash, created_at, expire_at, key_prefix) FROM api_key WHERE key_hash=$1`
	rows, err := r.Pool.Query(ctx, query, keyHash)
	defer rows.Close()

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if rows.Next() {
		resultKey := &domain.APIKey{}
		if err = rows.Scan(&resultKey.Key, &resultKey.CreatedAt, &resultKey.ExpiresAt, &resultKey.Prefix); err != nil {
			return nil, err
		}

		return resultKey, nil
	}

	return nil, errs.APIKeyNotFound
}

func (r *Repository) RevokeKeyByUserId(ctx context.Context, userId int64) error {
	log := appctx.GetLogger(ctx)

	query := `DELETE FROM api_key WHERE user_id=$1`
	_, err := r.Pool.Exec(ctx, query, userId)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (r *Repository) GetByKeyHash(ctx context.Context, hash string) (userId int64, key *domain.APIKey, err error) {
	log := appctx.GetLogger(ctx)
	query := `SELECT key_hash, user_id, expire_at, key_prefix FROM api_key WHERE key_hash=$1`
	rows, err := r.Pool.Query(ctx, query, hash)
	defer rows.Close()

	if err != nil {
		log.Error(err.Error())
		return 0, nil, err
	}

	if !rows.Next() {
		return 0, nil, nil
	}

	key = &domain.APIKey{}
	err = rows.Scan(&key.Key, &userId, &key.ExpiresAt, &key.Prefix)
	if err != nil {
		log.Error(err.Error())
		return 0, nil, err
	}

	log.Debug(fmt.Sprintf("Get key from db - %v", key))
	return userId, key, nil
}
