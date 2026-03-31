package appctx

import (
	"context"
	"errors"

	cfgutil "github.com/StewardMcCormick/Paste_Bin/config/cfg_util"
	"go.uber.org/zap"
)

type loggerCtxKey string
type envKey string
type userIdKey string

var (
	ErrInvalidEnv    = errors.New("incorrect value for env")
	ErrInvalidUserId = errors.New("incorrect value for user id")

	LoggerKey loggerCtxKey = "logger"
	EnvKey    envKey       = "env"
	UserIdKey userIdKey    = "user_id"
)

func WithLogger(parent context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(parent, LoggerKey, logger)
}

func GetLogger(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(LoggerKey).(*zap.Logger)
	if !ok {
		return zap.L()
	}

	return logger
}

func WithEnv(parent context.Context, env cfgutil.Env) context.Context {
	return context.WithValue(parent, EnvKey, env)
}

func GetEnv(ctx context.Context) (cfgutil.Env, error) {
	env, ok := ctx.Value(EnvKey).(cfgutil.Env)
	if !ok {
		return "", ErrInvalidEnv
	}

	return env, nil
}

func WithUserId(parent context.Context, userId int64) context.Context {
	return context.WithValue(parent, UserIdKey, userId)
}

func GetUserId(ctx context.Context) (int64, error) {
	id, ok := ctx.Value(UserIdKey).(int64)
	if !ok {
		return 0, ErrInvalidUserId
	}

	return id, nil
}
