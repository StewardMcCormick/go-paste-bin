package appctx

import (
	"context"
	"errors"
	cfgutil "github.com/StewardMcCormick/Paste_Bin/config/cfg_util"
	"go.uber.org/zap"
)

type loggerCtxKey string
type requestIdCtxKey string
type envKey string

var (
	InvalidRequestIdError = errors.New("incorrect value for request id")
	InvalidEnvError       = errors.New("incorrect value for env")

	LoggerKey    loggerCtxKey    = "logger"
	RequestIdKey requestIdCtxKey = "request_id"
	EnvKey       envKey          = "env"
)

func WithRequestId(parent context.Context, requestId string) context.Context {
	return context.WithValue(parent, RequestIdKey, requestId)
}

func GetRequestId(ctx context.Context) (string, error) {
	id, ok := ctx.Value(RequestIdKey).(string)
	if !ok {
		return "", InvalidRequestIdError
	}

	return id, nil
}

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

func GetEnv(ctx context.Context) (string, error) {
	env, ok := ctx.Value(LoggerKey).(string)
	if !ok {
		return "", InvalidEnvError
	}

	return env, nil
}
