package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type pgxUnitOfWorkFactory struct {
	pool *pgxpool.Pool
}

func NewUWFactory(pool *pgxpool.Pool) *pgxUnitOfWorkFactory {
	return &pgxUnitOfWorkFactory{pool: pool}
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
