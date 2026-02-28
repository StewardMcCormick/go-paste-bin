package paste

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	"github.com/StewardMcCormick/Paste_Bin/internal/usecase/paste/mocks"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UseCaseTestSuite struct {
	suite.Suite
	repo     *mocks.MockRepository
	valid    *mocks.MockValidator
	security *mocks.MockSecurity
	useCase  *UseCase
}

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}

func (s *UseCaseTestSuite) SetupTest() {
	s.repo = mocks.NewMockRepository(s.T())
	s.valid = mocks.NewMockValidator(s.T())
	s.security = mocks.NewMockSecurity(s.T())

	testCfg := Config{DefaultPasteExpiresTime: 7 * time.Hour}
	s.useCase = NewUseCase(testCfg, s.repo, s.valid, s.security)
}

func (s *UseCaseTestSuite) TestCreate_Success_CorrectlyExpireTimeSetting() {
	ctx := appctx.WithUserId(context.Background(), 1)
	now := time.Now()

	cases := []struct {
		name     string
		setup    func()
		value    *dto.PasteRequest
		expected *dto.PasteResponse
	}{
		{
			"Paste with unset ExpireAt field",
			func() {
				expectedPaste := &domain.Paste{
					Id:        1,
					UserId:    1,
					Hash:      "some_hash",
					Privacy:   domain.PublicPolicy,
					Content:   "content",
					CreatedAt: now,
					ExpireAt:  now.Add(s.useCase.cfg.DefaultPasteExpiresTime),
				}

				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()
				s.security.EXPECT().
					GeneratePasteHash().
					Return("some_hash", nil).
					Once()
				s.repo.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(paste *domain.Paste) bool {
						return !paste.ExpireAt.IsZero() && paste.UserId == 1 && paste.Hash == "some_hash"
					})).
					Return(expectedPaste, nil).
					Once()
			},
			&dto.PasteRequest{
				Content: "content",
				Privacy: string(domain.PublicPolicy),
			},
			&dto.PasteResponse{
				Id:        1,
				Privacy:   string(domain.PublicPolicy),
				CreatedAt: now,
				ExpireAt:  now.Add(s.useCase.cfg.DefaultPasteExpiresTime),
				Content:   "content",
			},
		},
		{
			"Paste with set ExpireAt field",
			func() {
				expectedPaste := &domain.Paste{
					Id:        1,
					UserId:    1,
					Hash:      "some_hash",
					Privacy:   domain.PublicPolicy,
					Content:   "content",
					CreatedAt: now,
					ExpireAt:  now.Add(24 * 3 * time.Hour),
				}

				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()
				s.security.EXPECT().
					GeneratePasteHash().
					Return("some_hash", nil).
					Once()
				s.repo.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(paste *domain.Paste) bool {
						return !paste.ExpireAt.IsZero() && paste.UserId == 1 && paste.Hash == "some_hash"
					})).
					Return(expectedPaste, nil).
					Once()
			},
			&dto.PasteRequest{
				Content:  "content",
				Privacy:  string(domain.PublicPolicy),
				ExpireAt: now.Add(24 * 3 * time.Hour),
			},
			&dto.PasteResponse{
				Id:        1,
				Privacy:   string(domain.PublicPolicy),
				CreatedAt: now,
				ExpireAt:  now.Add(24 * 3 * time.Hour),
				Content:   "content",
			},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			tc.setup()

			result, err := s.useCase.Create(ctx, &dto.PasteRequest{
				Privacy: string(domain.PublicPolicy),
				Content: "content",
			})

			s.NoError(err)
			s.NotNil(result)
			s.Equal(tc.expected.Id, result.Id)
			s.Equal(tc.expected.Views, result.Views)
			s.Equal(tc.expected.Privacy, result.Privacy)
			s.Equal(tc.expected.CreatedAt, result.CreatedAt)
			s.Equal(tc.expected.ExpireAt, result.ExpireAt)
			s.Equal(tc.expected.Content, result.Content)
		})
	}
}

func (s *UseCaseTestSuite) TestCreate_Success_CorrectlyPasswordSetting() {
	ctx := appctx.WithUserId(context.Background(), 1)
	now := time.Now()

	cases := []struct {
		name     string
		setup    func()
		value    *dto.PasteRequest
		expected *dto.PasteResponse
	}{
		{
			"Public paste - password should be empty",
			func() {
				expectedPaste := &domain.Paste{
					Id:        1,
					UserId:    1,
					Hash:      "some_hash",
					Privacy:   domain.PublicPolicy,
					Content:   "content",
					CreatedAt: now,
					ExpireAt:  now.Add(s.useCase.cfg.DefaultPasteExpiresTime),
				}

				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()
				s.security.EXPECT().
					GeneratePasteHash().
					Return("some_hash", nil).
					Once()
				s.repo.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(paste *domain.Paste) bool {
						return paste.PasswordHash == "" && paste.Hash == "some_hash"
					})).
					Return(expectedPaste, nil).
					Once()
			},
			&dto.PasteRequest{
				Content: "content",
				Privacy: string(domain.PublicPolicy),
			},
			&dto.PasteResponse{
				Id:        1,
				Privacy:   string(domain.PublicPolicy),
				CreatedAt: now,
				ExpireAt:  now.Add(s.useCase.cfg.DefaultPasteExpiresTime),
				Content:   "content",
			},
		},
		{
			"Protected paste - password shouldn`t be empty",
			func() {
				expectedPaste := &domain.Paste{
					Id:        1,
					UserId:    1,
					Hash:      "some_hash",
					Privacy:   domain.ProtectedPolicy,
					Content:   "content",
					CreatedAt: now,
					ExpireAt:  now.Add(s.useCase.cfg.DefaultPasteExpiresTime),
				}

				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()
				s.security.EXPECT().
					GeneratePasteHash().
					Return("some_hash", nil).
					Once()
				s.security.EXPECT().
					HashPassword(mock.Anything).
					Return("password_hash", nil).
					Once()
				s.repo.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(paste *domain.Paste) bool {
						return paste.PasswordHash == "password_hash" && paste.Hash == "some_hash"
					})).
					Return(expectedPaste, nil).
					Once()
			},
			&dto.PasteRequest{
				Content:  "content",
				Privacy:  string(domain.ProtectedPolicy),
				Password: "password",
			},
			&dto.PasteResponse{
				Id:        1,
				Privacy:   string(domain.ProtectedPolicy),
				CreatedAt: now,
				ExpireAt:  now.Add(s.useCase.cfg.DefaultPasteExpiresTime),
				Content:   "content",
			},
		},
		{
			"Private paste - password should be empty",
			func() {
				expectedPaste := &domain.Paste{
					Id:        1,
					UserId:    1,
					Hash:      "some_hash",
					Privacy:   domain.PrivatePolicy,
					Content:   "content",
					CreatedAt: now,
					ExpireAt:  now.Add(s.useCase.cfg.DefaultPasteExpiresTime),
				}

				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()
				s.security.EXPECT().
					GeneratePasteHash().
					Return("some_hash", nil).
					Once()
				s.repo.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(paste *domain.Paste) bool {
						return paste.PasswordHash == "" && paste.Hash == "some_hash"
					})).
					Return(expectedPaste, nil).
					Once()
			},
			&dto.PasteRequest{
				Content: "content",
				Privacy: string(domain.PrivatePolicy),
			},
			&dto.PasteResponse{
				Id:        1,
				Privacy:   string(domain.PrivatePolicy),
				CreatedAt: now,
				ExpireAt:  now.Add(s.useCase.cfg.DefaultPasteExpiresTime),
				Content:   "content",
			},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			tc.setup()

			result, err := s.useCase.Create(ctx, tc.value)

			s.NoError(err)
			s.NotNil(result)
			s.Equal(tc.expected.Id, result.Id)
			s.Equal(tc.expected.Views, result.Views)
			s.Equal(tc.expected.Privacy, result.Privacy)
			s.Equal(tc.expected.CreatedAt, result.CreatedAt)
			s.Equal(tc.expected.ExpireAt, result.ExpireAt)
			s.Equal(tc.expected.Content, result.Content)
		})
	}
}

func (s *UseCaseTestSuite) TestCreate_Error() {
	var ctx context.Context

	cases := []struct {
		name     string
		setup    func()
		value    *dto.PasteRequest
		expected error
	}{
		{
			"Paste validation error",
			func() {
				ctx = appctx.WithUserId(context.Background(), 1)
				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(errs.ValidationProcessError).
					Once()
			},
			&dto.PasteRequest{},
			errs.ValidationProcessError,
		},
		{
			"Incorrect UserId from context",
			func() {
				ctx = context.WithValue(context.Background(), appctx.UserIdKey, "abc")
				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()
			},
			&dto.PasteRequest{},
			errs.InternalError,
		},
		{
			"Password hashing error on Protected Paste",
			func() {
				ctx = appctx.WithUserId(context.Background(), 1)
				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()
				s.security.EXPECT().
					HashPassword(mock.Anything).
					Return("", errors.New("password hashing error")).
					Once()
			},
			&dto.PasteRequest{Privacy: string(domain.ProtectedPolicy)},
			errs.InternalError,
		},
		{
			"Generate Paste hash error",
			func() {
				ctx = appctx.WithUserId(context.Background(), 1)
				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()
				s.security.EXPECT().
					GeneratePasteHash().
					Return("", errors.New("hashing error")).
					Once()
			},
			&dto.PasteRequest{},
			errs.InternalError,
		},
		{
			"Repo error",
			func() {
				ctx = appctx.WithUserId(context.Background(), 1)
				s.valid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()
				s.security.EXPECT().
					GeneratePasteHash().
					Return("hash", nil).
					Once()
				s.repo.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(nil, errors.New("db error")).
					Once()
			},
			&dto.PasteRequest{},
			errs.InternalError,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			tc.setup()

			result, err := s.useCase.Create(ctx, tc.value)

			s.Nil(result)
			s.Error(err)
			s.ErrorIs(err, tc.expected)
		})
	}
}
