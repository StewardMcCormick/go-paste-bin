package user

import (
	"context"
	"fmt"
	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	"github.com/StewardMcCormick/Paste_Bin/internal/util"
	"github.com/StewardMcCormick/Paste_Bin/internal/validation"
	"time"
)

type Config struct {
	APIKeyExpireDuration time.Duration `yaml:"api_key_expires_time" env-default:"168h"`
}

type Repository interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	Exists(ctx context.Context, username string) (bool, error)
}

type SecurityUtil interface {
	HashPassword(password string) (string, error)
	HashAPIKey(key string) string
	GenerateAPIKey(ctx context.Context) (keyPrefix string, key string, err error)
}

type UseCase struct {
	repo         Repository
	securityUtil SecurityUtil
	valid        *validation.UserValidator
	cfg          Config
}

func NewUseCase(repo Repository, securityUtil SecurityUtil, valid *validation.UserValidator, cfg Config) *UseCase {
	return &UseCase{
		repo:         repo,
		securityUtil: securityUtil,
		valid:        valid,
		cfg:          cfg,
	}
}

func (uc *UseCase) Registration(ctx context.Context, user *dto.CreateUserRequest) (*dto.UserResponse, error) {
	log := util.GetLoggerFromCtx(ctx)

	if err := uc.valid.Validate(user); err != nil {
		return nil, err
	}

	exists, err := uc.repo.Exists(ctx, user.Username)
	if err != nil {
		return nil, fmt.Errorf("%w - database error", errs.InternalError)
	}
	if exists {
		return nil, errs.UserAlreadyExists
	}

	hashedPass, err := uc.securityUtil.HashPassword(user.Password)
	if err != nil {
		return nil, fmt.Errorf("%w - hashing password error", errs.InternalError)
	}
	prefix, apiKey, err := uc.securityUtil.GenerateAPIKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w - generate API Key error", errs.InternalError)
	}
	hashedKey := uc.securityUtil.HashAPIKey(apiKey)

	createdUser, err := uc.repo.CreateUser(
		ctx, &domain.User{
			Username: user.Username,
			Password: hashedPass,
			APIKey: domain.APIKey{
				Key:       hashedKey,
				Prefix:    prefix,
				ExpiresAt: time.Now().Add(uc.cfg.APIKeyExpireDuration)},
			CreatedAt: time.Now(),
		},
	)

	if err != nil {
		log.Error(err.Error())
		return nil, fmt.Errorf("%w - user registration error", errs.InternalError)
	}

	createdUser.APIKey.Key = apiKey
	return createdUser.ToResponse(), nil
}
