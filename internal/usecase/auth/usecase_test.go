package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	cfgUtil "github.com/StewardMcCormick/Paste_Bin/config/cfg_util"
	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	repoMocks "github.com/StewardMcCormick/Paste_Bin/internal/repository/mocks"
	ucMocks "github.com/StewardMcCormick/Paste_Bin/internal/usecase/auth/mocks"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UseCaseTestSuite struct {
	suite.Suite
	uowFactory *ucMocks.MockUnitOfWorkFactory
	noTxUow    *repoMocks.MockNoTxUnitOfWork
	txUow      *repoMocks.MockTxUnitOfWork
	apiKeyRepo *repoMocks.MockAPIKeyRepository
	userRepo   *repoMocks.MockUserRepository
	security   *ucMocks.MockSecurity
	valid      *ucMocks.MockValidator
	useCase    *UseCase
}

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}

func (s *UseCaseTestSuite) SetupTest() {
	s.uowFactory = ucMocks.NewMockUnitOfWorkFactory(s.T())
	s.noTxUow = repoMocks.NewMockNoTxUnitOfWork(s.T())
	s.txUow = repoMocks.NewMockTxUnitOfWork(s.T())
	s.apiKeyRepo = repoMocks.NewMockAPIKeyRepository(s.T())
	s.userRepo = repoMocks.NewMockUserRepository(s.T())

	s.security = ucMocks.NewMockSecurity(s.T())
	s.valid = ucMocks.NewMockValidator(s.T())

	s.useCase = NewUseCase(s.uowFactory, s.security,
		s.valid,
		Config{APIKeyExpireDuration: 162 * time.Hour},
	)

	s.setupExpectsForUow()
}

func (s *UseCaseTestSuite) setupExpectsForUow() {
	s.txUow.EXPECT().
		UserRepository().
		Return(s.userRepo).
		Maybe()

	s.txUow.EXPECT().
		APIKeyRepository().
		Return(s.apiKeyRepo).
		Maybe()

	s.noTxUow.EXPECT().
		UserRepository().
		Return(s.userRepo).
		Maybe()

	s.noTxUow.EXPECT().
		APIKeyRepository().
		Return(s.apiKeyRepo).
		Maybe()
}

func (s *UseCaseTestSuite) expectBeginTx() {
	s.uowFactory.EXPECT().
		Begin(mock.Anything).
		Return(s.txUow, nil).
		Once()
}

func (s *UseCaseTestSuite) expectCommitTx() {
	s.txUow.EXPECT().
		Commit(mock.Anything).
		Return(nil).
		Once()
}

func (s *UseCaseTestSuite) expectRollbackTx() {
	s.txUow.EXPECT().
		Rollback(mock.Anything).
		Return().
		Once()
}

func (s *UseCaseTestSuite) expectTx() {
	s.expectBeginTx()
	s.expectRollbackTx()
	s.expectCommitTx()
}

func (s *UseCaseTestSuite) Test_Registration_Success() {
	now := time.Now()
	expectedUser := &domain.User{
		Id:        0,
		Username:  "test",
		Password:  "test_pass",
		CreatedAt: now,
	}

	expectedApiKey := &domain.APIKey{
		Key:       "hashed_key",
		Prefix:    "pb_test",
		ExpiresAt: now.Add(s.useCase.cfg.APIKeyExpireDuration),
	}

	s.expectTx()

	s.valid.EXPECT().
		Validate(mock.Anything).
		Return(nil).
		Once()

	s.userRepo.EXPECT().
		GetByUsername(mock.Anything, mock.Anything).
		Return(nil, nil).
		Once()

	s.security.EXPECT().
		HashPassword(mock.Anything).
		Return("hashed_pass", nil).
		Once()

	s.security.EXPECT().
		GenerateAPIKey(mock.Anything).
		Return("pb_test", "pb_test_api-key", nil).
		Once()

	s.security.EXPECT().
		HashAPIKey(mock.Anything).
		Return("hashed_key").
		Once()

	s.userRepo.EXPECT().
		Create(mock.Anything, mock.Anything).
		Return(expectedUser, nil).
		Once()

	s.apiKeyRepo.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedApiKey, nil).
		Once()

	ctx := context.WithValue(context.Background(), appctx.EnvKey, cfgUtil.DevelopmentEnv)
	result, err := s.useCase.Registration(ctx,
		&dto.UserRequest{Username: expectedUser.Username, Password: expectedUser.Password})

	s.NoError(err)
	s.Equal(expectedUser.Id, result.Id)
	s.Equal(expectedUser.Username, result.Username)
	s.Equal(expectedApiKey.Key, result.APIKey.Key)
	s.Equal(expectedApiKey.ExpiresAt, result.APIKey.ExpiresAt)
	s.True(strings.HasPrefix(result.APIKey.Key, expectedApiKey.Prefix))
	s.NotNil(result.CreatedAt)
}

func (s *UseCaseTestSuite) Test_Registration_Error() {
	cases := []struct {
		name       string
		setupMocks func()
		wantError  error
	}{
		{
			"Validation error",
			func() {
				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(errs.ValidationProcessError).
					Once()
			},
			errs.ValidationProcessError,
		},
		{
			"Begin tx error",
			func() {
				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.uowFactory.EXPECT().
					Begin(mock.Anything).
					Return(nil, errors.New("begin tx error")).
					Once()
			},
			errs.InternalError,
		},
		{
			"Check user existing - Already Exists Error",
			func() {
				s.expectBeginTx()
				s.expectRollbackTx()

				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.userRepo.EXPECT().
					GetByUsername(mock.Anything, mock.Anything).
					Return(&domain.User{}, nil).
					Once()

			},
			errs.UserAlreadyExists,
		},
		{
			"Check user existing - Internal Error",
			func() {
				s.expectBeginTx()
				s.expectRollbackTx()

				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.userRepo.EXPECT().
					GetByUsername(mock.Anything, mock.Anything).
					Return(nil, errors.New("db error")).
					Once()

			},
			errs.InternalError,
		},
		{
			"Hashing password error",
			func() {
				s.expectBeginTx()
				s.expectRollbackTx()

				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.userRepo.EXPECT().
					GetByUsername(mock.Anything, mock.Anything).
					Return(nil, nil).
					Once()

				s.security.EXPECT().
					HashPassword(mock.Anything).
					Return("", errors.New("hashing error")).
					Once()

			},
			errs.InternalError,
		},
		{
			"Generate API Key error",
			func() {
				s.expectBeginTx()
				s.expectRollbackTx()

				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.userRepo.EXPECT().
					GetByUsername(mock.Anything, mock.Anything).
					Return(nil, nil).
					Once()

				s.security.EXPECT().
					HashPassword(mock.Anything).
					Return("hash", nil).
					Once()

				s.security.EXPECT().
					GenerateAPIKey(mock.Anything).
					Return("", "", errors.New("generate API Key error")).
					Once()

			},
			errs.InternalError,
		},
		{
			"User create error",
			func() {
				s.expectBeginTx()
				s.expectRollbackTx()

				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.userRepo.EXPECT().
					GetByUsername(mock.Anything, mock.Anything).
					Return(nil, nil).
					Once()

				s.security.EXPECT().
					HashPassword(mock.Anything).
					Return("hash", nil).
					Once()

				s.security.EXPECT().
					GenerateAPIKey(mock.Anything).
					Return("pb_test", "key", nil).
					Once()

				s.security.EXPECT().
					HashAPIKey(mock.Anything).
					Return("key_hash").
					Once()

				s.userRepo.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(nil, errors.New("user error")).
					Once()

			},
			errs.InternalError,
		},
		{
			"API key create error",
			func() {
				s.expectBeginTx()
				s.expectRollbackTx()

				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.userRepo.EXPECT().
					GetByUsername(mock.Anything, mock.Anything).
					Return(nil, nil).
					Once()

				s.security.EXPECT().
					HashPassword(mock.Anything).
					Return("hash", nil).
					Once()

				s.security.EXPECT().
					GenerateAPIKey(mock.Anything).
					Return("pb_test", "key", nil).
					Once()

				s.security.EXPECT().
					HashAPIKey(mock.Anything).
					Return("key_hash").
					Once()

				s.userRepo.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(&domain.User{}, nil).
					Once()

				s.apiKeyRepo.EXPECT().
					Create(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("API key error")).
					Once()

			},
			errs.InternalError,
		},
		{
			"Commit Error",
			func() {
				s.expectBeginTx()
				s.expectRollbackTx()

				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.userRepo.EXPECT().
					GetByUsername(mock.Anything, mock.Anything).
					Return(nil, nil).
					Once()

				s.security.EXPECT().
					HashPassword(mock.Anything).
					Return("hash", nil).
					Once()

				s.security.EXPECT().
					GenerateAPIKey(mock.Anything).
					Return("pb_test", "key", nil).
					Once()

				s.security.EXPECT().
					HashAPIKey(mock.Anything).
					Return("key_hash").
					Once()

				s.userRepo.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(&domain.User{}, nil).
					Once()

				s.apiKeyRepo.EXPECT().
					Create(mock.Anything, mock.Anything, mock.Anything).
					Return(&domain.APIKey{}, nil).
					Once()

				s.txUow.EXPECT().
					Commit(mock.Anything).
					Return(errors.New("commit error"))
			},
			errs.InternalError,
		},
	}

	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			s.SetupTest()
			tc.setupMocks()
			result, err := s.useCase.Registration(context.Background(),
				&dto.UserRequest{Username: "test", Password: "test_pass"},
			)

			s.Nil(result)
			s.NotNil(err)
			s.ErrorIs(err, tc.wantError)
		})
	}
}

func (s *UseCaseTestSuite) Test_Login_Success() {
	now := time.Now()
	expectedAPIKey := &dto.APIKeyResponse{
		Key:       "pb_test_test_api_key",
		ExpiresAt: now.Add(s.useCase.cfg.APIKeyExpireDuration),
	}

	s.expectTx()

	s.valid.EXPECT().
		Validate(mock.Anything).
		Return(nil).
		Once()

	s.userRepo.EXPECT().
		GetByUsername(mock.Anything, mock.Anything).
		Return(&domain.User{}, nil).
		Once()

	s.security.EXPECT().
		CompareHashAndPassword(mock.Anything, mock.Anything).
		Return(true).
		Once()

	s.apiKeyRepo.EXPECT().
		RevokeKeyByUserId(mock.Anything, mock.Anything).
		Return(nil).
		Once()

	s.security.EXPECT().
		GenerateAPIKey(mock.Anything).
		Return("pb_test", expectedAPIKey.Key, nil).
		Once()

	s.security.EXPECT().
		HashAPIKey(mock.Anything).
		Return("hashed_key").
		Once()

	s.apiKeyRepo.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(&domain.APIKey{Key: expectedAPIKey.Key, ExpiresAt: expectedAPIKey.ExpiresAt}, nil).
		Once()

	key, err := s.useCase.Login(context.Background(), &dto.UserRequest{})
	s.NoError(err)
	s.NotNil(key)
	s.Equal(expectedAPIKey.Key, key.Key)
	s.Equal(expectedAPIKey.ExpiresAt, key.ExpiresAt)
}

func (s *UseCaseTestSuite) Test_Login_Error() {
	cases := []struct {
		name       string
		setupMocks func()
		wantError  error
	}{
		{
			"Validation Error",
			func() {
				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(errs.ValidationProcessError).
					Once()
			},
			errs.ValidationProcessError,
		},
		{
			"Begin tx error",
			func() {
				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.uowFactory.EXPECT().
					Begin(mock.Anything).
					Return(nil, errors.New("begin tx error")).
					Once()
			},
			errs.InternalError,
		},
		{
			"Check user existing - user not found error",
			func() {
				s.expectBeginTx()
				s.expectRollbackTx()
				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.userRepo.EXPECT().
					GetByUsername(mock.Anything, mock.Anything).
					Return(nil, nil).
					Once()
			},
			errs.UserNotFound,
		},
		{
			"Check user existing - internal error",
			func() {
				s.expectBeginTx()
				s.expectRollbackTx()
				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.userRepo.EXPECT().
					GetByUsername(mock.Anything, mock.Anything).
					Return(nil, errors.New("db error")).
					Once()
			},
			errs.InternalError,
		},
		{
			"Incorrect password",
			func() {
				s.expectBeginTx()
				s.expectRollbackTx()
				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.userRepo.EXPECT().
					GetByUsername(mock.Anything, mock.Anything).
					Return(&domain.User{}, nil).
					Once()

				s.security.EXPECT().
					CompareHashAndPassword(mock.Anything, mock.Anything).
					Return(false).
					Once()
			},
			errs.Unauthorized,
		},
		{
			"Revoke key - internal error",
			func() {
				s.expectBeginTx()
				s.expectRollbackTx()
				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.userRepo.EXPECT().
					GetByUsername(mock.Anything, mock.Anything).
					Return(&domain.User{}, nil).
					Once()

				s.security.EXPECT().
					CompareHashAndPassword(mock.Anything, mock.Anything).
					Return(true).
					Once()

				s.apiKeyRepo.EXPECT().
					RevokeKeyByUserId(mock.Anything, mock.Anything).
					Return(errors.New("db error")).
					Once()
			},
			errs.InternalError,
		},
		{
			"Generate API Key error",
			func() {
				s.expectBeginTx()
				s.expectRollbackTx()
				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.userRepo.EXPECT().
					GetByUsername(mock.Anything, mock.Anything).
					Return(&domain.User{}, nil).
					Once()

				s.security.EXPECT().
					CompareHashAndPassword(mock.Anything, mock.Anything).
					Return(true).
					Once()

				s.apiKeyRepo.EXPECT().
					RevokeKeyByUserId(mock.Anything, mock.Anything).
					Return(nil).
					Once()

				s.security.EXPECT().
					GenerateAPIKey(mock.Anything).
					Return("", "", errors.New("API key error")).
					Once()
			},
			errs.InternalError,
		},
		{
			"Create API Key error",
			func() {
				s.expectBeginTx()
				s.expectRollbackTx()
				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.userRepo.EXPECT().
					GetByUsername(mock.Anything, mock.Anything).
					Return(&domain.User{}, nil).
					Once()

				s.security.EXPECT().
					CompareHashAndPassword(mock.Anything, mock.Anything).
					Return(true).
					Once()

				s.apiKeyRepo.EXPECT().
					RevokeKeyByUserId(mock.Anything, mock.Anything).
					Return(nil).
					Once()

				s.security.EXPECT().
					GenerateAPIKey(mock.Anything).
					Return("pb_test", "pb_test_test_key", nil).
					Once()

				s.security.EXPECT().
					HashAPIKey(mock.Anything).
					Return("key_hash").
					Once()

				s.apiKeyRepo.EXPECT().
					Create(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("db error")).
					Once()
			},
			errs.InternalError,
		},
		{
			"Create API Key error",
			func() {
				s.expectBeginTx()
				s.expectRollbackTx()
				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.userRepo.EXPECT().
					GetByUsername(mock.Anything, mock.Anything).
					Return(&domain.User{}, nil).
					Once()

				s.security.EXPECT().
					CompareHashAndPassword(mock.Anything, mock.Anything).
					Return(true).
					Once()

				s.apiKeyRepo.EXPECT().
					RevokeKeyByUserId(mock.Anything, mock.Anything).
					Return(nil).
					Once()

				s.security.EXPECT().
					GenerateAPIKey(mock.Anything).
					Return("pb_test", "pb_test_test_key", nil).
					Once()

				s.security.EXPECT().
					HashAPIKey(mock.Anything).
					Return("key_hash").
					Once()

				s.apiKeyRepo.EXPECT().
					Create(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("db error")).
					Once()
			},
			errs.InternalError,
		},
		{
			"Commit tx error",
			func() {
				s.expectBeginTx()
				s.expectRollbackTx()
				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.userRepo.EXPECT().
					GetByUsername(mock.Anything, mock.Anything).
					Return(&domain.User{}, nil).
					Once()

				s.security.EXPECT().
					CompareHashAndPassword(mock.Anything, mock.Anything).
					Return(true).
					Once()

				s.apiKeyRepo.EXPECT().
					RevokeKeyByUserId(mock.Anything, mock.Anything).
					Return(nil).
					Once()

				s.security.EXPECT().
					GenerateAPIKey(mock.Anything).
					Return("pb_test", "pb_test_test_key", nil).
					Once()

				s.security.EXPECT().
					HashAPIKey(mock.Anything).
					Return("key_hash").
					Once()

				s.apiKeyRepo.EXPECT().
					Create(mock.Anything, mock.Anything, mock.Anything).
					Return(&domain.APIKey{}, nil).
					Once()

				s.txUow.EXPECT().
					Commit(mock.Anything).
					Return(errors.New("commit error")).
					Once()
			},
			errs.InternalError,
		},
	}

	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			s.SetupTest()
			tc.setupMocks()

			key, err := s.useCase.Login(context.Background(), &dto.UserRequest{})

			s.Error(err)
			s.Nil(key)
			s.ErrorIs(err, tc.wantError)
		})
	}
}

func (s *UseCaseTestSuite) Test_Authenticate_Success() {
	var expectedId int64 = 1
	s.security.EXPECT().
		HashAPIKey(mock.Anything).
		Return("hash").
		Once()

	s.uowFactory.EXPECT().
		Exec(mock.Anything).
		Return(s.noTxUow).
		Once()

	s.apiKeyRepo.EXPECT().
		GetByKeyHash(mock.Anything, mock.Anything).
		Return(expectedId, &domain.APIKey{ExpiresAt: time.Now().Add(time.Hour)}, nil).
		Once()

	userId, err := s.useCase.Authenticate(context.Background(), "key")
	s.NoError(err)
	s.Equal(expectedId, userId)
}

func (s *UseCaseTestSuite) Test_Authenticate_Error() {
	cases := []struct {
		name       string
		setupMocks func()
		wantErr    error
	}{
		{
			"DB error - internal error",
			func() {
				s.security.EXPECT().
					HashAPIKey(mock.Anything).
					Return("hash").
					Once()

				s.uowFactory.EXPECT().
					Exec(mock.Anything).
					Return(s.noTxUow).
					Once()

				s.apiKeyRepo.EXPECT().
					GetByKeyHash(mock.Anything, mock.Anything).
					Return(0, nil, errors.New("db error")).
					Once()
			},
			errs.InternalError,
		},
		{
			"Key not found",
			func() {
				s.security.EXPECT().
					HashAPIKey(mock.Anything).
					Return("hash").
					Once()

				s.uowFactory.EXPECT().
					Exec(mock.Anything).
					Return(s.noTxUow).
					Once()

				s.apiKeyRepo.EXPECT().
					GetByKeyHash(mock.Anything, mock.Anything).
					Return(0, nil, nil).
					Once()
			},
			errs.Unauthorized,
		},
		{
			"Key expired",
			func() {
				s.security.EXPECT().
					HashAPIKey(mock.Anything).
					Return("hash").
					Once()

				s.uowFactory.EXPECT().
					Exec(mock.Anything).
					Return(s.noTxUow).
					Once()

				s.apiKeyRepo.EXPECT().
					GetByKeyHash(mock.Anything, mock.Anything).
					Return(0, &domain.APIKey{ExpiresAt: time.Now().Add(-1 * time.Hour)}, nil).
					Once()
			},
			errs.Unauthorized,
		},
	}

	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			s.SetupTest()
			tc.setupMocks()

			id, err := s.useCase.Authenticate(context.Background(), "key")

			s.Equal(int64(0), id)
			s.Error(err)
			s.ErrorIs(err, tc.wantErr)
		})
	}
}
