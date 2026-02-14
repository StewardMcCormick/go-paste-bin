package config

import (
	"github.com/StewardMcCormick/Paste_Bin/pkg/httpserver"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type App struct {
	Name    string `yaml:"name" env-required:"true"`
	Version string `yaml:"version" env-required:"true"`
}

type Config struct {
	App    App
	Server httpserver.Config
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
