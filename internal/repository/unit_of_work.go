package repository

import (
	"context"

	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	"github.com/StewardMcCormick/Paste_Bin/internal/repository/api_key"
	"github.com/StewardMcCormick/Paste_Bin/internal/repository/paste"
	"github.com/StewardMcCormick/Paste_Bin/internal/repository/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
}

type APIKeyRepository interface {
	Create(ctx context.Context, userId int64, key *domain.APIKey) (*domain.APIKey, error)
	RevokeKeyByUserId(ctx context.Context, userId int64) error
	GetByKeyHash(ctx context.Context, hash string) (userId int64, key *domain.APIKey, err error)
}

type PasteRepository interface {
}

type TxUnitOfWork interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context)

	UserRepository() UserRepository
	APIKeyRepository() APIKeyRepository
	PasteRepository() PasteRepository
}

type NoTxUnitOfWork interface {
	UserRepository() UserRepository
	APIKeyRepository() APIKeyRepository
	PasteRepository() PasteRepository
}

type pgxUnitOfWorkNoTx struct {
	pool *pgxpool.Pool
}

func (uw *pgxUnitOfWorkNoTx) UserRepository() UserRepository {
	return &user.Repository{Pool: uw.pool}
}

func (uw *pgxUnitOfWorkNoTx) APIKeyRepository() APIKeyRepository {
	return &api_key.Repository{Pool: uw.pool}
}

func (uw *pgxUnitOfWorkNoTx) PasteRepository() PasteRepository {
	return &paste.Repository{Pool: uw.pool}
}

type pgxUnitOfWorkTX struct {
	tx pgx.Tx
}

func (uwt *pgxUnitOfWorkTX) Commit(ctx context.Context) error {
	return uwt.tx.Commit(ctx)
}

func (uwt *pgxUnitOfWorkTX) Rollback(ctx context.Context) {
	uwt.tx.Rollback(ctx)
}

func (uwt *pgxUnitOfWorkTX) UserRepository() UserRepository {
	return &user.Repository{Pool: uwt.tx}
}

func (uwt *pgxUnitOfWorkTX) APIKeyRepository() APIKeyRepository {
	return &api_key.Repository{Pool: uwt.tx}
}

func (uwt *pgxUnitOfWorkTX) PasteRepository() PasteRepository {
	return &paste.Repository{Pool: uwt.tx}
}
