package paste

import (
	"context"
	"fmt"

	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
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

	return paste, nil
}
