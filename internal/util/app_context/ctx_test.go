package appctx

import (
	"context"
	"testing"

	cfgutil "github.com/StewardMcCormick/Paste_Bin/config/cfg_util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestLogger_NoError(t *testing.T) {
	log := &zap.Logger{}

	ctx := context.Background()
	ctx = WithLogger(ctx, log)

	ctxLog := GetLogger(ctx)
	assert.Equal(t, log, ctxLog)
}

func TestGetLogger_InvalidLoggerInCtx(t *testing.T) {
	log := "not a zap logger"

	ctx := context.WithValue(context.Background(), LoggerKey, log)

	ctxLogger := GetLogger(ctx)

	assert.Equal(t, zap.L(), ctxLogger)
}

func TestEnv_NoError(t *testing.T) {
	cases := []struct {
		name  string
		value cfgutil.Env
	}{
		{
			"Dev Env",
			cfgutil.DevelopmentEnv,
		},
		{
			"Prod env",
			cfgutil.ProductionEnv,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := WithEnv(context.Background(), tc.value)

			ctxEnv, err := GetEnv(ctx)

			assert.NoError(t, err)
			assert.Equal(t, tc.value, ctxEnv)
		})
	}
}

func TestEnv_Error(t *testing.T) {
	env := "invalid env"

	ctx := context.WithValue(context.Background(), EnvKey, env)

	ctxEnv, err := GetEnv(ctx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidEnv)
	assert.Equal(t, cfgutil.Env(""), ctxEnv)
}

func TestUserId_NoError(t *testing.T) {
	var id int64 = 1
	ctx := WithUserId(context.Background(), id)

	ctxId, err := GetUserId(ctx)

	assert.NoError(t, err)
	assert.Equal(t, id, ctxId)
}

func TestUserId_Error(t *testing.T) {
	id := "incorrect id"
	ctx := context.WithValue(context.Background(), UserIdKey, id)

	ctxId, err := GetUserId(ctx)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidUserId)
	assert.Equal(t, int64(0), ctxId)
}
