package user

import (
	"context"

	"github.com/StewardMcCormick/Paste_Bin/internal/adapter/postgres"
	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
)

type Repository struct {
	Pool postgres.DBTX
}

func NewRepository(pool postgres.DBTX) *Repository {
	return &Repository{Pool: pool}
}

func (r *Repository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	log := appctx.GetLogger(ctx)
	resultUser := &domain.User{}

	userQuery := `INSERT INTO users(username, password_hash, created_at) 
    				VALUES ($1, $2, $3) RETURNING *`
	err := r.Pool.QueryRow(ctx, userQuery, user.Username, user.Password, user.CreatedAt).
		Scan(&resultUser.Id, &resultUser.Username, &resultUser.Password, &resultUser.CreatedAt)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return resultUser, nil
}

func (r *Repository) Exists(ctx context.Context, username string) (bool, error) {
	log := appctx.GetLogger(ctx)

	query := `SELECT * FROM users WHERE username = $1`
	rows, err := r.Pool.Query(ctx, query, username)
	defer rows.Close()

	if err != nil {
		log.Error(err.Error())
		return false, err
	}

	if rows.Next() {
		return true, nil
	}

	return false, nil
}

func (r *Repository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	log := appctx.GetLogger(ctx)

	query := `SELECT id, username, password_hash, created_at FROM users WHERE username = $1`
	rows, err := r.Pool.Query(ctx, query, username)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	if rows.Next() {
		resultUser := &domain.User{}

		if err = rows.Scan(&resultUser.Id, &resultUser.Username, &resultUser.Password, &resultUser.CreatedAt); err != nil {
			log.Error(err.Error())
			return nil, err
		}

		return resultUser, nil
	}

	return nil, nil
}
