package logging

import (
	"github.com/StewardMcCormick/Paste_Bin/config/cfg_util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Level   string   `yaml:"level" env-default:"info"`
	Outputs []string `yaml:"outputs" env-required:"true"`
}

func NewLogger(cfg Config, env cfgutil.Env, appName, appVersion string) (*zap.Logger, error) {
	var loggerConfig zap.Config

	switch env {
	case cfgutil.DevelopmentEnv:
		loggerConfig = zap.NewDevelopmentConfig()
	default:
		loggerConfig = zap.NewProductionConfig()
	}

	var loggingLevel zapcore.Level
	if err := loggingLevel.Set(cfg.Level); err != nil {
		loggingLevel = zapcore.InfoLevel
	}

	loggerConfig.Level = zap.NewAtomicLevelAt(loggingLevel)

	loggerConfig.OutputPaths = cfg.Outputs

	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, err
	}

	logger = logger.With(
		zap.String("App-Name", appName),
		zap.String("App-Version", appVersion),
	)

	return logger, nil
}
