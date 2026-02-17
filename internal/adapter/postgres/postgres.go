package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type ConnectionConfig struct {
	ConnectionTimeout     time.Duration `yaml:"connection_timeout" default:"5s"`
	MaxConnections        int32         `yaml:"max_connections" default:"4"`
	MinConnections        int32         `yaml:"min_connections" default:"0"`
	MaxConnectionLiveTime time.Duration `yaml:"max_connection_live_time" default:"1h"`
	MaxConnectionIdleTime time.Duration `yaml:"max_connection_idle_time" default:"30m"`
	HealthCheckPeriod     time.Duration `yaml:"health_check_period" default:"1m"`
}

type Config struct {
	DbUrl            string
	User             string           `env:"POSTGRES_USER" required:"true"`
	Password         string           `env:"POSTGRES_PASSWORD" required:"true"`
	Port             string           `env:"POSTGRES_PORT" env-default:"5432"`
	Host             string           `env:"POSTGRES_HOST" env-default:"localhost"`
	DbName           string           `env:"POSTGRES_DB" required:"true"`
	ConnectionConfig ConnectionConfig `yaml:"connection_config"`
}

type Pool struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, cfg *Config) (*Pool, error) {
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DbName,
	)

	config, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		return nil, err
	}
	cfg.DbUrl = dbUrl

	config.ConnConfig.ConnectTimeout = cfg.ConnectionConfig.ConnectionTimeout
	config.MaxConns = cfg.ConnectionConfig.MaxConnections
	config.MinConns = cfg.ConnectionConfig.MinConnections
	config.MaxConnLifetime = cfg.ConnectionConfig.MaxConnectionLiveTime
	config.MaxConnIdleTime = cfg.ConnectionConfig.MaxConnectionIdleTime
	config.HealthCheckPeriod = cfg.ConnectionConfig.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return &Pool{pool}, nil
}

func (p *Pool) Close() {
	p.pool.Close()
}
