package paste

import (
	"context"

	"github.com/StewardMcCormick/Paste_Bin/internal/repository"
)

type UnitOfWorkFactory interface {
	Exec(ctx context.Context) repository.NoTxUnitOfWork
	Begin(ctx context.Context) (repository.TxUnitOfWork, error)
}

type UseCase struct {
	uow UnitOfWorkFactory
}

func NewUseCase(uow UnitOfWorkFactory) *UseCase {
	return &UseCase{uow: uow}
}
