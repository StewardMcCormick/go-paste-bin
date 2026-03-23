package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
	"go.uber.org/zap"
)

func (uc *UseCase) Authenticate(ctx context.Context, apiKey string) (userId int64, err error) {
	log := appctx.GetLogger(ctx)
	hash := uc.securityUtil.HashAPIKey(apiKey)

	key, err := uc.uow.Exec(ctx).APIKeyRepository().GetByKeyHash(ctx, hash)
	if err != nil {
		if errors.Is(err, errs.APIKeyNotFound) {
			log.Info(fmt.Sprintf("api-key not found - %s", err))
			return 0, fmt.Errorf("api-key not found - %w", errs.Unauthorized)
		}

		return 0, fmt.Errorf("%w - find key error", errs.InternalError)
	}
	if key == nil || key.ExpiresAt.Compare(time.Now()) <= 0 {
		log.Debug(fmt.Sprintf("Key from DB - %v", key))
		log.Info("authentication failed")
		return 0, fmt.Errorf("%w - key invalid or expired - you should get a new key", errs.Unauthorized)
	}

	log.Info(
		"new authenticate",
		zap.Int64("user_id", key.UserId),
	)

	return key.UserId, nil
}
