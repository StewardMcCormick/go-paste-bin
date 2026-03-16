package paste

import (
	"context"
	"errors"
	"fmt"

	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Cache interface {
	Set(ctx context.Context, hash string, content *domain.PasteContent)
	Get(ctx context.Context, hash string) *domain.PasteContent
	RevokeByKey(ctx context.Context, key string)
}

type Repository struct {
	pool  *pgxpool.Pool
	Cache Cache
}

func NewRepository(pool *pgxpool.Pool, cache Cache) *Repository {
	return &Repository{pool: pool, Cache: cache}
}

func (r *Repository) Create(ctx context.Context, paste *domain.Paste) (*domain.Paste, error) {
	log := appctx.GetLogger(ctx)
	tx, err := r.pool.Begin(ctx)
	defer tx.Rollback(ctx)

	if err != nil {
		log.Error(fmt.Sprintf("%s - tx begin error", err.Error()))
		return nil, err
	}

	query := `INSERT INTO paste_info(user_id, paste_hash, views, privacy, password_hash, created_at, expire_at) 
				VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	err = tx.QueryRow(ctx, query,
		paste.UserId, paste.Hash, 0, paste.Privacy, paste.PasswordHash, paste.CreatedAt, paste.ExpireAt,
	).Scan(&paste.Id)

	if err != nil {
		log.Error(fmt.Sprintf("%s - paste-info saving error", err.Error()))
		return nil, err
	}

	query = `INSERT INTO paste_content(paste_id, content) VALUES ($1, $2)`

	_, err = tx.Exec(ctx, query, paste.Id, paste.Content)
	if err != nil {
		log.Error(fmt.Sprintf("%s - paste-content saving error", err.Error()))
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error(fmt.Sprintf("%s - tx commit error", err.Error()))
		return nil, err
	}

	log.Debug(fmt.Sprintf("new paste: id - %d, user_id - %d", paste.Id, paste.UserId))
	return paste, nil
}

func (r *Repository) GetByHash(ctx context.Context, hash string) (*domain.Paste, error) {
	log := appctx.GetLogger(ctx)

	if content := r.Cache.Get(ctx, hash); content != nil {
		paste, err := r.getInfoByHash(ctx, hash)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				log.Debug(fmt.Sprintf("paste %s not found", hash))
				return nil, errs.PasteNotFound
			}
			log.Error(fmt.Sprintf("%v - getting paste error", err))
			return nil, errs.InternalError
		}

		paste.Content = *content

		log.Info(fmt.Sprintf("get paste - %s", hash))
		return paste, nil
	}

	paste, err := r.getFullPasteByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Debug(fmt.Sprintf("paste %s not found", hash))
			return nil, errs.PasteNotFound
		}
		log.Error(fmt.Sprintf("%v - getting paste error", err))
		return nil, errs.InternalError
	}

	log.Info(fmt.Sprintf("get paste - %s", hash))

	log.Debug("update Cache")
	r.Cache.Set(ctx, hash, &paste.Content)

	return paste, nil
}

func (r *Repository) getInfoByHash(ctx context.Context, hash string) (*domain.Paste, error) {
	log := appctx.GetLogger(ctx)

	query := `SELECT pi.id, pi.user_id, pi.views, pi.privacy, pi.password_hash, pi.created_at, pi.expire_at, pi.paste_hash 
				FROM paste_info pi WHERE paste_hash=$1`

	result := &domain.Paste{}

	log.Info("new DB query")
	err := r.pool.QueryRow(ctx, query, hash).
		Scan(
			&result.Id,
			&result.UserId,
			&result.Views,
			&result.Privacy,
			&result.PasswordHash,
			&result.CreatedAt,
			&result.ExpireAt,
			&result.Hash,
		)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Repository) getFullPasteByHash(ctx context.Context, hash string) (*domain.Paste, error) {
	log := appctx.GetLogger(ctx)

	query := `SELECT pi.id, pi.user_id, pi.views, pi.privacy, pi.password_hash, pi.created_at, pi.expire_at, pi.paste_hash, pc.content
    			FROM paste_info pi JOIN paste_content pc ON pi.id = pc.paste_id 
				WHERE paste_hash=$1`

	result := &domain.Paste{}

	log.Info("new DB query")
	err := r.pool.QueryRow(ctx, query, hash).
		Scan(
			&result.Id,
			&result.UserId,
			&result.Views,
			&result.Privacy,
			&result.PasswordHash,
			&result.CreatedAt,
			&result.ExpireAt,
			&result.Hash,
			&result.Content,
		)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Repository) Update(ctx context.Context, paste *domain.Paste) (*domain.Paste, error) {
	log := appctx.GetLogger(ctx)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		log.Error(fmt.Sprintf("tx begin error - %v", err))
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `UPDATE paste_info SET privacy=$1, password_hash=$2, expire_at=$3 WHERE paste_hash=$4 RETURNING id`

	id := int64(0)
	err = tx.QueryRow(ctx, query, paste.Privacy, paste.PasswordHash, paste.ExpireAt, paste.Hash).Scan(&id)
	if err != nil {
		log.Error(fmt.Sprintf("update paste-info error - %v", err))
		return nil, err
	}

	query = `UPDATE paste_content SET content=$1 WHERE paste_id=$2`

	_, err = tx.Exec(ctx, query, paste.Content, id)
	if err != nil {
		log.Error(fmt.Sprintf("update paste-content error - %v", err))
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error(fmt.Sprintf("commit tx error - %v", err))
		return nil, err
	}

	r.Cache.RevokeByKey(ctx, paste.Hash)

	return paste, nil
}
