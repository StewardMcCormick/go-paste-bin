package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	apiKeyCacheName = "API_KEY_CACHE"
	pasteCacheName  = "PASTE_CACHE"
	ipRateName      = "IP_RATE_LIMITING"
	userIdRateName  = "USER_ID_RATE_LIMITING"
)

type CacheConfig struct {
	Host        string        `env:"REDIS_CACHE_HOST" required:"true"`
	Port        int           `env:"REDIS_CACHE_PORT" required:"true"`
	PoolSize    int           `yaml:"pool_size" env-default:"10"`
	PoolTimeout time.Duration `yaml:"pool_timeout" env-default:"5s"`
	Password    string        `env:"REDIS_CACHE_PASSWORD" required:"true"`
	Db          int
}

type RateConfig struct {
	Host        string        `env:"REDIS_RATE_HOST" required:"true"`
	Port        int           `env:"REDIS_RATE_PORT" required:"true"`
	PoolSize    int           `yaml:"pool_size" env-default:"10"`
	PoolTimeout time.Duration `yaml:"pool_timeout" env-default:"5s"`
	Password    string        `env:"REDIS_RATE_PASSWORD" required:"true"`
	Db          int
}

type Config struct {
	Cache *CacheConfig `yaml:"cache"`
	Rate  *RateConfig  `yaml:"rate"`
}

type Manager struct {
	clients map[string]*redis.Client
	mu      *sync.Mutex
}

func NewManager(cfg Config) (*Manager, error) {
	manager := &Manager{
		clients: make(map[string]*redis.Client, 2),
		mu:      &sync.Mutex{},
	}

	if cfg.Cache != nil {
		cfg.Cache.Db = 0
		apiKeyCacheClient, err := newClient(apiKeyCacheName, cfg.Cache)
		if err != nil {
			return nil, err
		}

		cfg.Cache.Db = 1
		pasteCacheClient, err := newClient(pasteCacheName, cfg.Cache)
		if err != nil {
			return nil, err
		}

		manager.clients[apiKeyCacheName] = apiKeyCacheClient
		manager.clients[pasteCacheName] = pasteCacheClient
	}

	if cfg.Rate != nil {
		cfg.Rate.Db = 0
		ipRateClient, err := newClient(ipRateName, cfg.Rate)
		if err != nil {
			return nil, err
		}

		cfg.Rate.Db = 1
		userIdRateClient, err := newClient(userIdRateName, cfg.Rate)
		if err != nil {
			return nil, err
		}

		manager.clients[ipRateName] = ipRateClient
		manager.clients[userIdRateName] = userIdRateClient
	}

	return manager, nil
}

func newClient[T RedisConfig](name string, cfg T) (*redis.Client, error) {

	addr := fmt.Sprintf("%s:%d", cfg.GetHost(), cfg.GetPort())

	client := redis.NewClient(&redis.Options{
		Addr:        addr,
		Password:    cfg.GetPassword(),
		PoolSize:    cfg.GetPoolSize(),
		PoolTimeout: cfg.GetPoolTimeout(),
		DB:          cfg.GetDb(),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis client init error: %s, error: %w", name, err)
	}

	return client, nil
}

func (m *Manager) GetAPIKeyCacheClient() *redis.Client {
	return m.clients[apiKeyCacheName]
}

func (m *Manager) GetPasteCacheClient() *redis.Client {
	return m.clients[pasteCacheName]
}

func (m *Manager) GetIpRateLimitingClient() *redis.Client {
	return m.clients[ipRateName]
}

func (m *Manager) GetUserIdRateLimitingClient() *redis.Client {
	return m.clients[userIdRateName]
}

func (m *Manager) Close() error {
	var err error
	for _, client := range m.clients {
		err = client.Close()
	}

	return err
}
