package user

import (
	"context"

	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
)

type AuthUseCase interface {
	Registration(ctx context.Context, user *dto.UserRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, user *dto.UserRequest) (*dto.APIKeyResponse, error)
}

type httpHandlers struct {
	authUseCase AuthUseCase
}

func NewHandler(authUseCase AuthUseCase) *httpHandlers {
	return &httpHandlers{
		authUseCase: authUseCase,
	}
}
