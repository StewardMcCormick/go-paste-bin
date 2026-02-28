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

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{pool: pool}
}

func (r *repository) Create(ctx context.Context, paste *domain.Paste) (*domain.Paste, error) {
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

func (r *repository) GetByHash(ctx context.Context, hash string) (*domain.Paste, error) {
	log := appctx.GetLogger(ctx)

	query := `SELECT pi.id, pi.user_id, pi.views, pi.privacy, pi.created_at, pi.expire_at, pc.content
    			FROM paste_info pi JOIN paste_content pc ON pi.id = pc.paste_id 
				WHERE paste_hash=$1`

	result := &domain.Paste{}
	err := r.pool.QueryRow(ctx, query, hash).
		Scan(
			&result.Id,
			&result.UserId,
			&result.Views,
			&result.Privacy,
			&result.CreatedAt,
			&result.ExpireAt,
			&result.Content,
		)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Debug(fmt.Sprintf("paste %s not found", hash))
			return nil, errs.PasteNotFound
		}
		log.Error(fmt.Sprintf("%v - getting paste error", err))
		return nil, err
	}

	log.Debug(fmt.Sprintf("get past: hash - %s", hash))
	return result, nil
}
