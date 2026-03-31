package test

import (
	"context"
	"testing"
	"time"

	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	apikey "github.com/StewardMcCormick/Paste_Bin/internal/repository/api_key"
	appcache "github.com/StewardMcCormick/Paste_Bin/internal/repository/cache"
	"github.com/stretchr/testify/suite"
)

type APIKeyRepoIntTest struct {
	suite.Suite
	repo *apikey.Repository
}

func TestAPIKeyIntTest(t *testing.T) {
	suite.Run(t, new(APIKeyRepoIntTest))
}

func (s *APIKeyRepoIntTest) SetupSuite() {
	cache := appcache.NewAPIKeyCache(apiKeyCacheRedisClient)

	s.repo = apikey.NewRepository(pool, cache)

	createTestUser(context.Background(), pool)
	createTestAPIKey(context.Background(), pool)
}

func (s *APIKeyRepoIntTest) TearDownSuite() {
	query := `TRUNCATE TABLE users CASCADE`

	_, err := pool.Exec(context.Background(), query)
	s.Require().NoError(err)
}

func (s *APIKeyRepoIntTest) Test_GetByHash_Success() {
	resultFromRepo, err := s.repo.GetByKeyHash(context.Background(), testAPIKey.Key)
	s.Require().NoError(err)

	s.NotNil(resultFromRepo)
	s.Equal(testAPIKey.Key, resultFromRepo.Key)
	s.Equal(testAPIKey.Prefix, resultFromRepo.Prefix)
	s.Equal(testAPIKey.UserId, resultFromRepo.UserId)
	s.True(testAPIKey.CreatedAt.Equal(resultFromRepo.CreatedAt))
	s.True(testAPIKey.ExpiresAt.Equal(resultFromRepo.ExpiresAt))

	query := `SELECT * FROM api_key WHERE key_hash=$1`

	resultFromDb := &domain.APIKey{}
	err = pool.QueryRow(context.Background(), query, testAPIKey.Key).Scan(
		&resultFromDb.Key,
		&resultFromDb.UserId,
		&resultFromDb.CreatedAt,
		&resultFromDb.ExpiresAt,
		&resultFromDb.Prefix,
	)
	s.Require().NoError(err)

	s.NotNil(resultFromRepo)
	s.Equal(resultFromDb.Key, resultFromRepo.Key)
	s.Equal(resultFromDb.Prefix, resultFromRepo.Prefix)
	s.Equal(resultFromDb.UserId, resultFromRepo.UserId)
	s.True(resultFromDb.CreatedAt.Equal(resultFromRepo.CreatedAt))
	s.True(resultFromDb.ExpiresAt.Equal(resultFromRepo.ExpiresAt))

	keyFromCache := s.repo.Cache.Get(context.Background(), testAPIKey.Key)

	s.NotNil(keyFromCache)

	s.Equal(testAPIKey.Key, keyFromCache.Key)
	s.Equal(testAPIKey.Prefix, keyFromCache.Prefix)
	s.Equal(testAPIKey.UserId, keyFromCache.UserId)
	s.True(testAPIKey.CreatedAt.Equal(keyFromCache.CreatedAt))
	s.True(testAPIKey.ExpiresAt.Equal(keyFromCache.ExpiresAt))
}

func (s *APIKeyRepoIntTest) Test_GetByHash_NotFound() {
	result, err := s.repo.GetByKeyHash(context.Background(), "not_exist")

	s.Nil(result)
	s.ErrorIs(err, errs.ErrAPIKeyNotFound)
}

func (s *APIKeyRepoIntTest) Test_Create_Success() {
	key := &domain.APIKey{
		UserId:    testUser.Id,
		Key:       "key",
		Prefix:    "pb_test",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	result, err := s.repo.Create(context.Background(), key.UserId, key)

	s.NoError(err)
	s.NotNil(result)

	s.Equal(key.UserId, result.UserId)
	s.Equal(key.Key, result.Key)
	s.Equal(key.Prefix, result.Prefix)
	s.True(key.CreatedAt.Equal(result.CreatedAt))
	s.True(key.ExpiresAt.Equal(result.ExpiresAt))

	keyFromDb, err := s.repo.GetByKeyHash(context.Background(), key.Key)

	s.NoError(err)
	s.NotNil(keyFromDb)

	s.Equal(keyFromDb.UserId, result.UserId)
	s.Equal(keyFromDb.Key, result.Key)
	s.Equal(keyFromDb.Prefix, result.Prefix)
	s.True(keyFromDb.CreatedAt.Equal(result.CreatedAt))
	s.True(keyFromDb.ExpiresAt.Equal(result.ExpiresAt))
}

func (s *APIKeyRepoIntTest) Test_Create_AlreadyExist() {
	result, err := s.repo.Create(context.Background(), testAPIKey.UserId, testAPIKey)

	s.Nil(result)
	s.Error(err)
}

func (s *APIKeyRepoIntTest) Test_Create_WithIncorrectUserId() {
	result, err := s.repo.Create(context.Background(), 100, testAPIKey)

	s.Nil(result)
	s.Error(err)
}

func (s *APIKeyRepoIntTest) Test_RevokeByUserId_Success() {
	err := s.repo.RevokeKeyByUserId(context.Background(), testAPIKey.UserId)

	s.NoError(err)

	time.Sleep(time.Second)
	key, err := s.repo.GetByKeyHash(context.Background(), testAPIKey.Key)

	s.Error(err)
	s.Nil(key)
}

func (s *APIKeyRepoIntTest) Test_RevokeByUserId_NotFound() {
	err := s.repo.RevokeKeyByUserId(context.Background(), 100)

	s.NoError(err)
}
