package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	"github.com/StewardMcCormick/Paste_Bin/internal/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	postgresContainerName = "postgres"
	cacheContainerName    = "cache"
)

var (
	pool                   *pgxpool.Pool
	pasteCacheRedisClient  *redis.Client
	apiKeyCacheRedisClient *redis.Client

	testUser   = &domain.User{}
	testAPIKey = &domain.APIKey{}
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	composeStack, err := compose.NewDockerCompose("docker-compose-test.yaml")
	if err != nil {
		panic(err)
	}

	err = composeStack.WaitForService("postgres",
		wait.ForLog("database system is ready to accept connections").
			WithStartupTimeout(30*time.Second)).
		Up(ctx, compose.Wait(true))

	if err != nil {
		panic(err)
	}

	err = composeStack.WaitForService("cache",
		wait.ForLog("Ready to accept connections tcp").
			WithStartupTimeout(30*time.Second)).
		Up(ctx, compose.Wait(true))

	if err != nil {
		panic(err)
	}

	pool = initPgxPool(ctx, composeStack)
	pasteCacheRedisClient = initRedisClient(ctx, composeStack, 0)
	apiKeyCacheRedisClient = initRedisClient(ctx, composeStack, 1)

	m.Run()

	pool.Close()
	err = pasteCacheRedisClient.Close()
	if err != nil {
		panic(err)
	}
	err = apiKeyCacheRedisClient.Close()
	if err != nil {
		panic(err)
	}

	err = composeStack.Down(ctx, compose.RemoveImagesLocal)
	if err != nil {
		panic(err)
	}
}

func initRedisClient(ctx context.Context, stack compose.ComposeStack, dbNum int) *redis.Client {
	container, err := stack.ServiceContainer(ctx, cacheContainerName)
	if err != nil {
		panic(err)
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		panic(err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		panic(err)
	}

	addr := fmt.Sprintf("%s:%d", host, port.Int())

	client := redis.NewClient(&redis.Options{
		Addr:        addr,
		Password:    "test_password",
		PoolTimeout: 10 * time.Second,
		DB:          dbNum,
	})

	if err = client.Ping(ctx).Err(); err != nil {
		panic(err)
	}

	return client
}

func initPgxPool(ctx context.Context, compose compose.ComposeStack) *pgxpool.Pool {
	postgresConnString := getPostgresConnectionString(ctx, compose)
	pool, err := pgxpool.New(ctx, postgresConnString)
	if err != nil {
		panic(err)
	}

	if err = pool.Ping(ctx); err != nil {
		panic(err)
	}

	err = migrations.Exec(postgresConnString)
	if err != nil {
		panic(err)
	}

	return pool
}

func getPostgresConnectionString(ctx context.Context, compose compose.ComposeStack) string {
	container, err := compose.ServiceContainer(ctx, postgresContainerName)
	if err != nil {
		panic(err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		panic(err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		"test_user", "test_password", host, port.Int(), "test_db",
	)
}

func createTestUser(ctx context.Context, pool *pgxpool.Pool) {
	query := `INSERT INTO users(username, password_hash, created_at) VALUES (
    	'test_user', 'test_pass_hash', now()
	) RETURNING *`

	err := pool.QueryRow(context.Background(), query).Scan(
		&testUser.Id,
		&testUser.Username,
		&testUser.Password,
		&testUser.CreatedAt,
	)

	if err != nil {
		panic(err)
	}
}

func createTestAPIKey(ctx context.Context, pool *pgxpool.Pool) {
	query := `INSERT INTO api_key(key_hash, user_id, created_at, expire_at, key_prefix) VALUES (
    	'test_hash', $1, $2, $3, 'pb_test'                                                                              
	) RETURNING *`

	now := time.Now()
	err := pool.QueryRow(ctx, query, testUser.Id, now, now.Add(3*time.Hour)).Scan(
		&testAPIKey.Key,
		&testAPIKey.UserId,
		&testAPIKey.CreatedAt,
		&testAPIKey.ExpiresAt,
		&testAPIKey.Prefix,
	)

	if err != nil {
		panic(err)
	}
}
