package paste

import (
	"context"

	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	"github.com/StewardMcCormick/Paste_Bin/internal/handler"
)

type UseCase interface {
	Create(ctx context.Context, request *dto.PasteRequest) (*dto.PasteResponse, error)
	GetByHash(ctx context.Context, hash string) (*dto.PasteResponse, error)
}

type httpHandlers struct {
	cfg     handler.Config
	useCase UseCase
}

func NewHandlers(cfg handler.Config, useCase UseCase) *httpHandlers {
	return &httpHandlers{cfg: cfg, useCase: useCase}
}
