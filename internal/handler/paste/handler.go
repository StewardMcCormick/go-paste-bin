package paste

import (
	"context"

	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
)

type UseCase interface {
	Create(ctx context.Context, request *dto.PasteRequest) (*dto.PasteResponse, error)
}

type httpHandlers struct {
	useCase UseCase
}

func NewHandlers(useCase UseCase) *httpHandlers {
	return &httpHandlers{useCase: useCase}
}
