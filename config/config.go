package config

import (
	"github.com/StewardMcCormick/Paste_Bin/config/cfg_util"
	"github.com/StewardMcCormick/Paste_Bin/internal/adapter/postgres"
	"github.com/StewardMcCormick/Paste_Bin/internal/usecase/auth"
	"github.com/StewardMcCormick/Paste_Bin/internal/usecase/paste"
	"github.com/StewardMcCormick/Paste_Bin/pkg/httpserver"
	"github.com/StewardMcCormick/Paste_Bin/pkg/logging"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type App struct {
	Name    string      `yaml:"name" env-required:"true"`
	Version string      `yaml:"version" env-required:"true"`
	Env     cfgutil.Env `yaml:"env" env-default:"prod"`
}

type Config struct {
	App      App
	Server   httpserver.Config
	Logger   logging.Config
	Postgres postgres.Config
	Auth     auth.Config
	Paste    paste.Config
}

func InitConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = cleanenv.ReadConfig("config.yaml", cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
