package paste

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
	views "github.com/StewardMcCormick/Paste_Bin/internal/util/views_worker"
)

type Config struct {
	DefaultPasteExpiresTime time.Duration `yaml:"default_paste_expires_time" evn-default:"168h"`
}

type Repository interface {
	Create(context.Context, *domain.Paste) (*domain.Paste, error)
	GetByHash(ctx context.Context, hash string) (*domain.Paste, error)
}

type Validator interface {
	Validate(request *dto.PasteRequest) error
}

type Security interface {
	HashPassword(password string) (string, error)
	CompareHashAndPassword(hash string, pass string) bool
	GeneratePasteHash() (string, error)
}

type ViewWorker interface {
	SendEvent(ctx context.Context, event views.ViewEvent)
}

type UseCase struct {
	cfg        Config
	repo       Repository
	valid      Validator
	security   Security
	viewWorker ViewWorker
}

func NewUseCase(cfg Config, repo Repository, valid Validator, security Security, viewWorker ViewWorker) *UseCase {
	return &UseCase{
		cfg:        cfg,
		repo:       repo,
		valid:      valid,
		security:   security,
		viewWorker: viewWorker,
	}
}

func (uc *UseCase) Create(ctx context.Context, request *dto.PasteRequest) (*dto.PasteResponse, error) {
	log := appctx.GetLogger(ctx)
	if err := uc.valid.Validate(request); err != nil {
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

func (uc *UseCase) GetByHash(ctx context.Context, request dto.GetPasteRequest, hash string) (*dto.PasteResponse, error) {
	log := appctx.GetLogger(ctx)

	paste, err := uc.repo.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, errs.PasteNotFound) {
			return nil, err
		}
		log.Error(fmt.Sprintf("%v - get paste error", err))
		return nil, fmt.Errorf("%w - get paste error", errs.InternalError)
	}

	userId, err := appctx.GetUserId(ctx)
	if err != nil {
		log.Error(fmt.Sprintf("%v - get user_id from ctx error", err))
		return nil, fmt.Errorf("%w - get user_id from ctx error", errs.InternalError)
	}

	if paste.Privacy == domain.PrivatePolicy && userId != paste.UserId {
		log.Debug(fmt.Sprintf("get paste Forbidden(Private): from - %d, to paste with user_id - %d", userId, paste.UserId))
		return nil, errs.Forbidden
	} else if paste.Privacy == domain.ProtectedPolicy &&
		!uc.security.CompareHashAndPassword(paste.PasswordHash, request.Password) {
		log.Debug(fmt.Sprintf("get paste Forbidden(Protected): from - %d, to paste with user_id - %d", userId, paste.UserId))
		return nil, errs.Unauthorized
	}

	paste.Views += 1
	uc.viewWorker.SendEvent(ctx, views.ViewEvent{PasteId: paste.Id})

	return paste.ToResponse(), nil
}
