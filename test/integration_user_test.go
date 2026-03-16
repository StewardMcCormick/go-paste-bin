package test

import (
	"context"
	"testing"
	"time"

	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	"github.com/StewardMcCormick/Paste_Bin/internal/repository/user"
	"github.com/stretchr/testify/suite"
)

type UserRepoIntTestSuite struct {
	suite.Suite
	repo *user.Repository
}

func TestUserRepoInt(t *testing.T) {
	suite.Run(t, new(UserRepoIntTestSuite))
}

func (s *UserRepoIntTestSuite) SetupSuite() {
	s.repo = user.NewRepository(pool)

	createTestUser(context.Background(), pool)
}

func (s *UserRepoIntTestSuite) TearDownSuite() {
	query := `TRUNCATE TABLE users CASCADE`

	_, err := pool.Exec(context.Background(), query)

	s.Require().NoError(err)
}

func (s *UserRepoIntTestSuite) Test_GetByUsername_Success() {
	resultFromRepo, err := s.repo.GetByUsername(context.Background(), testUser.Username)

	s.Require().NoError(err)

	s.Equal(testUser.Id, resultFromRepo.Id)
	s.Equal(testUser.Username, resultFromRepo.Username)
	s.Equal(testUser.Password, resultFromRepo.Password)
	s.True(testUser.CreatedAt.Equal(resultFromRepo.CreatedAt))
}

func (s *UserRepoIntTestSuite) Test_GetByUsername_NotFound() {
	result, err := s.repo.GetByUsername(context.Background(), "not_exist")

	s.NoError(err)
	s.Nil(result)
}

func (s *UserRepoIntTestSuite) Test_Exists_True() {
	result, err := s.repo.Exists(context.Background(), testUser.Username)

	s.NoError(err)
	s.True(result)
}

func (s *UserRepoIntTestSuite) Test_Exists_False() {
	result, err := s.repo.Exists(context.Background(), "not_exist")

	s.NoError(err)
	s.False(result)
}

func (s *UserRepoIntTestSuite) Test_Create_Success() {
	now := time.Now()
	value := &domain.User{
		Username:  "user_2",
		Password:  "pass_2",
		CreatedAt: now,
	}

	result, err := s.repo.Create(context.Background(), value)

	s.NoError(err)
	s.Equal(value.Username, result.Username)
	s.Equal(value.Password, result.Password)
	s.True(value.CreatedAt.Equal(result.CreatedAt))

	userFromRepo, err := s.repo.GetByUsername(context.Background(), value.Username)

	s.NoError(err)
	s.Equal(value.Username, userFromRepo.Username)
	s.Equal(value.Password, userFromRepo.Password)
	s.True(value.CreatedAt.Equal(userFromRepo.CreatedAt))
}

func (s *UserRepoIntTestSuite) Test_Create_AlreadyExists() {
	result, err := s.repo.Create(context.Background(), testUser)

	s.Nil(result)
	s.Error(err)
}
