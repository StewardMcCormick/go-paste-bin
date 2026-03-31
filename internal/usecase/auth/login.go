package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
	"go.uber.org/zap"
)

func (uc *UseCase) Login(ctx context.Context, user *dto.UserRequest) (*dto.APIKeyResponse, error) {
	log := appctx.GetLogger(ctx)
	if err := uc.valid.Validate(user); err != nil {
		log.Debug(fmt.Sprintf("%v - validation error", err))
		return nil, err
	}

	tx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w - tx beggining error", errs.ErrInternal)
	}
	defer tx.Rollback(ctx)

	userFromDb, err := tx.UserRepository().GetByUsername(ctx, user.Username)
	if err != nil {
		return nil, fmt.Errorf("%w - database error", errs.ErrInternal)
	}
	if userFromDb == nil {
		return nil, errs.ErrUserNotFound
	}
	if !uc.securityUtil.CompareHashAndPassword(userFromDb.Password, user.Password) {
		return nil, errs.ErrUnauthorized
	}

	err = tx.APIKeyRepository().RevokeKeyByUserId(ctx, userFromDb.Id)
	if err != nil {
		return nil, fmt.Errorf("%w - API key revoke error", errs.ErrInternal)
	}

	newKey, err := uc.generateNewKey(ctx)
	if err != nil {
		log.Error(fmt.Sprintf("%s - API key generation error", err.Error()))
		return nil, fmt.Errorf("%w - API key ganaration error", errs.ErrInternal)
	}

	newKeyFromDb, err := tx.APIKeyRepository().Create(
		ctx, userFromDb.Id,
		&domain.APIKey{
			Key:       newKey.Hash,
			Prefix:    newKey.Prefix,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(uc.cfg.APIKeyExpireDuration),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%w - API key creating error", errs.ErrInternal)
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error(err.Error())
		return nil, fmt.Errorf("%w - tx commit error", errs.ErrInternal)
	}

	newKeyFromDb.Key = newKey.Key

	log.Info(
		"user login",
		zap.String("username", user.Username),
	)

	return newKeyFromDb.ToResponse(), nil
}
