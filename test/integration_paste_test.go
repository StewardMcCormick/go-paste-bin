package test

import (
	"context"
	"testing"
	"time"

	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	appcache "github.com/StewardMcCormick/Paste_Bin/internal/repository/cache"
	"github.com/StewardMcCormick/Paste_Bin/internal/repository/paste"
	"github.com/stretchr/testify/suite"
)

type PasteRepoIntTest struct {
	suite.Suite

	repo *paste.Repository
}

func TestPasteRepoInt(t *testing.T) {
	suite.Run(t, new(PasteRepoIntTest))
}

func (s *PasteRepoIntTest) SetupSuite() {
	cache := appcache.NewPasteCache(pasteCacheRedisClient)

	s.repo = paste.NewRepository(pool, cache)

	createTestUser(context.Background(), pool)
	createTestAPIKey(context.Background(), pool)
	createTestPaste(context.Background(), pool)
}

func (s *PasteRepoIntTest) Test_GetByHash_Success() {
	result, err := s.repo.GetByHash(context.Background(), testPaste.Hash)

	s.NoError(err)
	s.Equal(testPaste.Id, result.Id)
	s.Equal(testPaste.Hash, result.Hash)
	s.Equal(testPaste.PasswordHash, result.PasswordHash)
	s.Equal(testPaste.Privacy, result.Privacy)
	s.Equal(testPaste.Views, result.Views)
	s.Equal(testPaste.Content, result.Content)
	s.Equal(testPaste.UserId, result.UserId)
	s.True(testPaste.CreatedAt.Equal(result.CreatedAt))
	s.True(testPaste.ExpireAt.Equal(result.ExpireAt))

	time.Sleep(time.Second)
	resultFromCache := s.repo.Cache.Get(context.Background(), testPaste.Hash)

	time.Sleep(time.Second)
	s.Require().NotNil(resultFromCache)
	s.Equal(&testPaste.Content, resultFromCache)
}

func (s *PasteRepoIntTest) Test_GetByHash_NotExist() {
	result, err := s.repo.GetByHash(context.Background(), "not exist")

	s.Nil(result)
	s.Error(err)
}

func (s *PasteRepoIntTest) Test_Create_Success() {
	pasteToSave := &domain.Paste{
		UserId:       testUser.Id,
		Hash:         "new hash",
		Views:        0,
		Privacy:      domain.PublicPolicy,
		PasswordHash: "pass",
		CreatedAt:    time.Now(),
		ExpireAt:     time.Now().Add(time.Hour),
		Content:      "content",
	}

	result, err := s.repo.Create(context.Background(), pasteToSave)

	s.Require().NoError(err)
	s.Require().NotNil(result)

	s.Equal(pasteToSave.UserId, result.UserId)
	s.Equal(pasteToSave.Hash, result.Hash)
	s.Equal(pasteToSave.Views, result.Views)
	s.Equal(pasteToSave.Privacy, result.Privacy)
	s.Equal(pasteToSave.PasswordHash, result.PasswordHash)
	s.Equal(pasteToSave.Content, result.Content)
	s.True(pasteToSave.CreatedAt.Equal(result.CreatedAt))
	s.True(pasteToSave.ExpireAt.Equal(result.ExpireAt))

	resultFromDb, err := s.repo.GetByHash(context.Background(), pasteToSave.Hash)

	s.Require().NoError(err)
	s.Require().NotNil(resultFromDb)

	s.Equal(pasteToSave.UserId, resultFromDb.UserId)
	s.Equal(pasteToSave.Hash, resultFromDb.Hash)
	s.Equal(pasteToSave.Views, resultFromDb.Views)
	s.Equal(pasteToSave.Privacy, resultFromDb.Privacy)
	s.Equal(pasteToSave.PasswordHash, resultFromDb.PasswordHash)
	s.Equal(pasteToSave.Content, resultFromDb.Content)
	s.True(pasteToSave.CreatedAt.Equal(resultFromDb.CreatedAt))
	s.True(pasteToSave.ExpireAt.Equal(resultFromDb.ExpireAt))
}

func (s *PasteRepoIntTest) Test_Create_IncorrectUserId() {
	pasteToSave := &domain.Paste{
		UserId:       100,
		Hash:         "new hash",
		Views:        0,
		Privacy:      domain.PublicPolicy,
		PasswordHash: "pass",
		CreatedAt:    time.Now(),
		ExpireAt:     time.Now().Add(time.Hour),
		Content:      "content",
	}

	result, err := s.repo.Create(context.Background(), pasteToSave)

	s.Require().Nil(result)
	s.Error(err)
}

func (s *PasteRepoIntTest) Test_Update_Success() {
	toUpdate := &domain.Paste{
		Id:           testPaste.Id,
		UserId:       testUser.Id,
		Hash:         testPaste.Hash,
		Views:        0,
		Privacy:      domain.PrivatePolicy,
		PasswordHash: "pass",
		CreatedAt:    testPaste.CreatedAt,
		ExpireAt:     time.Now().Add(2 * time.Hour),
		Content:      "new content",
	}

	result, err := s.repo.Update(context.Background(), toUpdate)

	s.Require().NoError(err)
	s.Require().NotNil(result)

	s.Equal(toUpdate.UserId, result.UserId)
	s.Equal(toUpdate.Hash, result.Hash)
	s.Equal(toUpdate.Views, result.Views)
	s.Equal(toUpdate.Privacy, result.Privacy)
	s.Equal(toUpdate.PasswordHash, result.PasswordHash)
	s.Equal(toUpdate.Content, result.Content)
	s.True(toUpdate.CreatedAt.Equal(result.CreatedAt))
	s.True(toUpdate.ExpireAt.Equal(result.ExpireAt))

	time.Sleep(time.Second)
	resultFromDb, err := s.repo.GetByHash(context.Background(), testPaste.Hash)

	s.Require().NoError(err)
	s.Require().NotNil(resultFromDb)

	s.Equal(toUpdate.UserId, resultFromDb.UserId)
	s.Equal(toUpdate.Hash, resultFromDb.Hash)
	s.Equal(toUpdate.Views, resultFromDb.Views)
	s.Equal(toUpdate.Privacy, resultFromDb.Privacy)
	s.Equal(toUpdate.PasswordHash, resultFromDb.PasswordHash)
	s.Equal(toUpdate.Content, resultFromDb.Content)
	s.True(toUpdate.CreatedAt.Equal(resultFromDb.CreatedAt))
	s.True(toUpdate.ExpireAt.Equal(resultFromDb.ExpireAt))
}
