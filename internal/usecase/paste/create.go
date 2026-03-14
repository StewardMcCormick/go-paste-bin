package paste

import (
	"context"
	"fmt"
	"time"

	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
)

func (uc *UseCase) Create(ctx context.Context, request *dto.PasteRequest) (*dto.PasteResponse, error) {
	log := appctx.GetLogger(ctx)
	if err := uc.createRequestValid.Validate(request); err != nil {
		log.Debug(fmt.Sprintf("validation error - %v", err))
		return nil, err
	}

	userId, err := appctx.GetUserId(ctx)
	if err != nil {
		log.Error(fmt.Sprintf("%v - user id parsing error", err))
		return nil, fmt.Errorf("%w - user id parsing error", errs.InternalError)
	}

	requestToDomain := &domain.Paste{
		UserId:    userId,
		Content:   domain.PasteContent(request.Content),
		Privacy:   domain.PrivacyPolicy(request.Privacy),
		CreatedAt: time.Now(),
	}

	if request.ExpireAt.IsZero() || request.ExpireAt.Before(time.Now()) {
		requestToDomain.ExpireAt = time.Now().Add(uc.cfg.DefaultPasteExpiresTime)
	}

	if requestToDomain.Privacy == domain.ProtectedPolicy {
		hashedPassword, err := uc.security.HashPassword(request.Password)
		if err != nil {
			log.Error(fmt.Sprintf("Hashing password error - %v", err))
			return nil, fmt.Errorf("%w - password hashing error", errs.InternalError)
		}

		requestToDomain.PasswordHash = hashedPassword
	}

	hash, err := uc.security.GeneratePasteHash()
	if err != nil {
		log.Error(fmt.Sprintf("%v - ganaration rand string error", err))
		return nil, fmt.Errorf("%w - ganaration rand string error", errs.InternalError)
	}
	requestToDomain.Hash = hash

	paste, err := uc.repo.Create(ctx, requestToDomain)
	if err != nil {
		log.Error(fmt.Sprintf("%v - Paste saving error", err))
		return nil, fmt.Errorf("%w - Paste saving error", errs.InternalError)
	}

	return paste.ToResponse(), nil
}
