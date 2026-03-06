package repository

import (
	"context"

	"github.com/StewardMcCormick/Paste_Bin/internal/repository/api_key"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgxUnitOfWorkFactory struct {
	pool     *pgxpool.Pool
	keyCache api_key.Cache
}

func NewUWFactory(pool *pgxpool.Pool, apiKeyCache api_key.Cache) *pgxUnitOfWorkFactory {
	return &pgxUnitOfWorkFactory{pool: pool, keyCache: apiKeyCache}
}

func (f *pgxUnitOfWorkFactory) Exec(ctx context.Context) NoTxUnitOfWork {
	return &pgxUnitOfWorkNoTx{pool: f.pool}
}

func (f *pgxUnitOfWorkFactory) Begin(ctx context.Context) (TxUnitOfWork, error) {
	tx, err := f.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return &pgxUnitOfWorkTX{tx: tx}, nil
}
