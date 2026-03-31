package paste

import (
	"context"
	"errors"
	"fmt"

	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
)

func (uc *UseCase) UpdatePaste(ctx context.Context, hash string, request *dto.UpdatePasteRequest) (*dto.PasteResponse, error) {
	log := appctx.GetLogger(ctx)

	if err := uc.updateRequestValid.Validate(request); err != nil {
		log.Debug(fmt.Sprintf("validation error - %v", err))
		return nil, fmt.Errorf("%w - paste validation error", err)
	}

	pasteFromDb, err := uc.repo.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, errs.ErrPasteNotFound) {
			return nil, err
		}
		log.Error(fmt.Sprintf("Get paste error - %v", err))
		return nil, fmt.Errorf("%w - get past error", errs.ErrInternal)
	}

	requestToDomain := &domain.Paste{}
	if request.Content == string(pasteFromDb.Content) || request.Content == "" {
		requestToDomain.Content = pasteFromDb.Content
	} else {
		requestToDomain.Content = domain.PasteContent(request.Content)
	}
	if request.Privacy == string(pasteFromDb.Privacy) || request.Privacy == "" {
		requestToDomain.Privacy = pasteFromDb.Privacy
	} else {
		requestToDomain.Privacy = domain.PrivacyPolicy(request.Privacy)
	}
	if request.Privacy != string(domain.ProtectedPolicy) {
		request.Password = ""
	} else if request.Password == "" ||
		uc.security.CompareHashAndPassword(pasteFromDb.PasswordHash, request.Password) {
		requestToDomain.PasswordHash = pasteFromDb.PasswordHash
	} else {
		passHash, err := uc.security.HashPassword(request.Password)
		if err != nil {
			log.Error(fmt.Sprintf("hash password error - %v", err))
			return nil, fmt.Errorf("%w - hash password error", errs.ErrInternal)
		}

		requestToDomain.PasswordHash = passHash
	}
	if request.ExpireAt.Equal(pasteFromDb.ExpireAt) || request.ExpireAt.IsZero() {
		requestToDomain.ExpireAt = pasteFromDb.ExpireAt
	} else {
		requestToDomain.ExpireAt = request.ExpireAt
	}

	requestToDomain.Hash = hash
	result, err := uc.repo.Update(ctx, requestToDomain)
	if err != nil {
		log.Error(fmt.Sprintf("Paste updating error - %v", err))
		return nil, fmt.Errorf("%w - paste update error", errs.ErrInternal)
	}

	result.Id = pasteFromDb.Id
	result.Views = pasteFromDb.Views
	result.CreatedAt = pasteFromDb.CreatedAt
	return result.ToResponse(), nil
}
