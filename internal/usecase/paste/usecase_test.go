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
	views "github.com/StewardMcCormick/Paste_Bin/internal/util/views_worker"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UseCaseTestSuite struct {
	suite.Suite
	repo               *mocks.MockRepository
	createRequestValid *mocks.MockCreateRequestValidator
	updateRequestValid *mocks.MockUpdateRequestValidator
	security           *mocks.MockSecurity
	worker             *mocks.MockViewWorker
	useCase            *UseCase
}

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}

func (s *UseCaseTestSuite) SetupTest() {
	s.repo = mocks.NewMockRepository(s.T())
	s.createRequestValid = mocks.NewMockCreateRequestValidator(s.T())
	s.updateRequestValid = mocks.NewMockUpdateRequestValidator(s.T())
	s.security = mocks.NewMockSecurity(s.T())
	s.worker = mocks.NewMockViewWorker(s.T())

	testCfg := Config{DefaultPasteExpiresTime: 7 * time.Hour}
	s.useCase = NewUseCase(testCfg, s.repo, s.createRequestValid, s.updateRequestValid, s.security, s.worker)
}

func (s *UseCaseTestSuite) Test_Create_Success_CorrectlyExpireTimeSetting() {
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

				s.createRequestValid.EXPECT().
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

				s.createRequestValid.EXPECT().
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

func (s *UseCaseTestSuite) Test_Create_Success_CorrectlyPasswordSetting() {
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

				s.createRequestValid.EXPECT().
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

				s.createRequestValid.EXPECT().
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

				s.createRequestValid.EXPECT().
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
			s.SetupTest()
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

func (s *UseCaseTestSuite) Test_Create_Error() {
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
				s.createRequestValid.EXPECT().
					Validate(mock.Anything).
					Return(errs.ErrValidationProcess).
					Once()
			},
			&dto.PasteRequest{},
			errs.ErrValidationProcess,
		},
		{
			"Incorrect UserId from context",
			func() {
				ctx = context.WithValue(context.Background(), appctx.UserIdKey, "abc")
				s.createRequestValid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()
			},
			&dto.PasteRequest{},
			errs.ErrInternal,
		},
		{
			"Password hashing error on Protected Paste",
			func() {
				ctx = appctx.WithUserId(context.Background(), 1)
				s.createRequestValid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()
				s.security.EXPECT().
					HashPassword(mock.Anything).
					Return("", errors.New("password hashing error")).
					Once()
			},
			&dto.PasteRequest{Privacy: string(domain.ProtectedPolicy)},
			errs.ErrInternal,
		},
		{
			"Generate Paste hash error",
			func() {
				ctx = appctx.WithUserId(context.Background(), 1)
				s.createRequestValid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()
				s.security.EXPECT().
					GeneratePasteHash().
					Return("", errors.New("hashing error")).
					Once()
			},
			&dto.PasteRequest{},
			errs.ErrInternal,
		},
		{
			"Repo error",
			func() {
				ctx = appctx.WithUserId(context.Background(), 1)
				s.createRequestValid.EXPECT().
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
			errs.ErrInternal,
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

func (s *UseCaseTestSuite) Test_GetByHash_Success() {
	ctx := appctx.WithUserId(context.Background(), 1)

	now := time.Now()
	afterWeek := now.Add(24 * 7 * time.Hour)

	cases := []struct {
		name     string
		setup    func()
		value    string
		expected *dto.PasteResponse
	}{
		{
			"Get Public paste Paste",
			func() {
				s.repo.EXPECT().
					GetByHash(mock.Anything, mock.Anything).
					Return(
						&domain.Paste{
							Id:        1,
							UserId:    1,
							Privacy:   domain.PublicPolicy,
							Hash:      "hash",
							CreatedAt: now,
							ExpireAt:  afterWeek,
						}, nil,
					).Once()
			},
			"hash",
			&dto.PasteResponse{
				Id:        1,
				Privacy:   string(domain.PublicPolicy),
				Hash:      "hash",
				CreatedAt: now,
				ExpireAt:  afterWeek,
			},
		},
		{
			"Get Private paste Paste",
			func() {
				s.repo.EXPECT().
					GetByHash(mock.Anything, mock.Anything).
					Return(
						&domain.Paste{
							Id:        1,
							UserId:    1,
							Privacy:   domain.PrivatePolicy,
							Hash:      "hash",
							Content:   "content",
							CreatedAt: now,
							ExpireAt:  afterWeek,
						}, nil,
					).Once()
			},
			"hash",
			&dto.PasteResponse{
				Id:        1,
				Privacy:   string(domain.PrivatePolicy),
				Hash:      "hash",
				Content:   "content",
				CreatedAt: now,
				ExpireAt:  afterWeek,
			},
		},
		{
			"Get Protected paste Paste",
			func() {
				s.repo.EXPECT().
					GetByHash(mock.Anything, mock.Anything).
					Return(
						&domain.Paste{
							Id:        1,
							UserId:    1,
							Privacy:   domain.ProtectedPolicy,
							Hash:      "hash",
							Content:   "content",
							CreatedAt: now,
							ExpireAt:  afterWeek,
						}, nil,
					).Once()
				s.security.EXPECT().
					CompareHashAndPassword(mock.Anything, mock.Anything).
					Return(true).
					Once()
			},
			"hash",
			&dto.PasteResponse{
				Id:        1,
				Privacy:   string(domain.ProtectedPolicy),
				Hash:      "hash",
				Content:   "content",
				CreatedAt: now,
				ExpireAt:  afterWeek,
			},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			s.worker.EXPECT().
				SendEvent(mock.Anything, mock.MatchedBy(func(event views.ViewEvent) bool {
					return event.PasteId == tc.expected.Id
				})).Once()

			res, err := s.useCase.GetByHash(ctx, dto.GetPasteRequest{Password: "pass"}, tc.value)

			s.NoError(err)
			s.NotNil(res)

			s.Equal(tc.expected.Id, res.Id)
			s.Equal(tc.expected.Privacy, res.Privacy)
			s.Equal(tc.expected.Hash, res.Hash)
			s.Equal(tc.expected.Content, res.Content)
			s.Equal(tc.expected.CreatedAt, res.CreatedAt)
			s.Equal(tc.expected.ExpireAt, res.ExpireAt)
		})
	}
}

func (s *UseCaseTestSuite) Test_GetByHash_Error() {
	var ctx context.Context

	cases := []struct {
		name     string
		value    string
		setup    func()
		expected error
	}{
		{
			"Repo error - Paste not found",
			"hash",
			func() {
				ctx = appctx.WithUserId(context.Background(), 1)
				s.repo.EXPECT().
					GetByHash(mock.Anything, mock.Anything).
					Return(nil, errs.ErrPasteNotFound).
					Once()
			},
			errs.ErrPasteNotFound,
		},
		{
			"Repo error - internal error",
			"hash",
			func() {
				ctx = appctx.WithUserId(context.Background(), 1)
				s.repo.EXPECT().
					GetByHash(mock.Anything, mock.Anything).
					Return(nil, errors.New("db error")).
					Once()
			},
			errs.ErrInternal,
		},
		{
			"Incorrect User_Id in ctx",
			"hash",
			func() {
				ctx = context.WithValue(context.Background(), appctx.UserIdKey, "invalid user_id")
				s.repo.EXPECT().
					GetByHash(mock.Anything, mock.Anything).
					Return(&domain.Paste{}, nil).
					Once()
			},
			errs.ErrInternal,
		},
		{
			"ErrForbidden error - get Private paste with another user_id",
			"hash",
			func() {
				ctx = appctx.WithUserId(context.Background(), 1)
				s.repo.EXPECT().
					GetByHash(mock.Anything, mock.Anything).
					Return(&domain.Paste{UserId: 0, Privacy: domain.PrivatePolicy}, nil).
					Once()

			},
			errs.ErrForbidden,
		},
		{
			"ErrUnauthorized error - wrong password",
			"hash",
			func() {
				ctx = appctx.WithUserId(context.Background(), 1)
				s.repo.EXPECT().
					GetByHash(mock.Anything, mock.Anything).
					Return(&domain.Paste{Privacy: domain.ProtectedPolicy}, nil).
					Once()
				s.security.EXPECT().
					CompareHashAndPassword(mock.Anything, mock.Anything).
					Return(false).
					Once()
			},
			errs.ErrUnauthorized,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			res, err := s.useCase.GetByHash(ctx, dto.GetPasteRequest{Password: "pass"}, tc.value)

			s.Nil(res)
			s.Error(err)
			s.ErrorIs(err, tc.expected)
		})
	}
}

func (s *UseCaseTestSuite) Test_UpdatePaste_Success() {
	now := time.Now()
	cases := []struct {
		name     string
		value    *dto.UpdatePasteRequest
		expected *dto.PasteResponse
		setup    func()
	}{
		{
			"Update all fields with no-protected privacy",
			&dto.UpdatePasteRequest{
				Content:  "new content",
				Privacy:  string(domain.PublicPolicy),
				Password: "pass",
				ExpireAt: now.Add(time.Hour),
			},
			&dto.PasteResponse{
				Id:        10,
				Views:     2,
				Privacy:   string(domain.PublicPolicy),
				CreatedAt: now,
				ExpireAt:  now.Add(time.Hour),
				Content:   "new content",
			},
			func() {
				s.updateRequestValid.EXPECT().
					Validate(&dto.UpdatePasteRequest{
						Content:  "new content",
						Privacy:  string(domain.PublicPolicy),
						Password: "pass",
						ExpireAt: now.Add(time.Hour),
					}).
					Return(nil).
					Once()

				s.repo.EXPECT().
					GetByHash(mock.Anything, "hash").
					Return(&domain.Paste{
						Id:        10,
						Views:     2,
						Privacy:   domain.PrivatePolicy,
						Hash:      "hash",
						CreatedAt: now,
						ExpireAt:  now.Add(2 * time.Hour),
						Content:   "content",
					}, nil).
					Once()

				s.repo.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(paste *domain.Paste) bool {
						return paste.Content == "new content" &&
							paste.Privacy == domain.PublicPolicy &&
							paste.PasswordHash == "" && paste.ExpireAt.Equal(now.Add(time.Hour))
					})).
					Return(&domain.Paste{
						Id:        10,
						Views:     2,
						Privacy:   domain.PublicPolicy,
						Hash:      "hash",
						CreatedAt: now,
						ExpireAt:  now.Add(time.Hour),
						Content:   "new content",
					}, nil).Once()
			},
		},
		{
			"Update all fields with protected privacy",
			&dto.UpdatePasteRequest{
				Content:  "new content",
				Privacy:  string(domain.ProtectedPolicy),
				Password: "new pass",
				ExpireAt: now.Add(time.Hour),
			},
			&dto.PasteResponse{
				Id:        10,
				Views:     2,
				Privacy:   string(domain.ProtectedPolicy),
				CreatedAt: now,
				ExpireAt:  now.Add(time.Hour),
				Content:   "new content",
			},
			func() {
				s.updateRequestValid.EXPECT().
					Validate(&dto.UpdatePasteRequest{
						Content:  "new content",
						Privacy:  string(domain.ProtectedPolicy),
						Password: "new pass",
						ExpireAt: now.Add(time.Hour),
					}).
					Return(nil).
					Once()

				s.repo.EXPECT().
					GetByHash(mock.Anything, "hash").
					Return(&domain.Paste{
						Id:           10,
						Views:        2,
						Privacy:      domain.PublicPolicy,
						Hash:         "hash",
						PasswordHash: "old hash",
						CreatedAt:    now,
						ExpireAt:     now.Add(2 * time.Hour),
						Content:      "content",
					}, nil).
					Once()

				s.security.EXPECT().
					CompareHashAndPassword("old hash", "new pass").
					Return(false).
					Once()

				s.security.EXPECT().
					HashPassword("new pass").
					Return("new hash", nil).
					Once()

				s.repo.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(paste *domain.Paste) bool {
						return paste.Content == "new content" &&
							paste.Privacy == domain.ProtectedPolicy &&
							paste.PasswordHash == "new hash" && paste.ExpireAt.Equal(now.Add(time.Hour))
					})).
					Return(&domain.Paste{
						Id:        10,
						Views:     2,
						Privacy:   domain.ProtectedPolicy,
						Hash:      "hash",
						CreatedAt: now,
						ExpireAt:  now.Add(time.Hour),
						Content:   "new content",
					}, nil).Once()
			},
		},
		{
			"No updated fields with no-protected privacy",
			&dto.UpdatePasteRequest{
				Content:  "content",
				Privacy:  string(domain.PublicPolicy),
				Password: "pass",
				ExpireAt: now.Add(2 * time.Hour),
			},
			&dto.PasteResponse{
				Id:        10,
				Views:     2,
				Privacy:   string(domain.PublicPolicy),
				CreatedAt: now,
				ExpireAt:  now.Add(2 * time.Hour),
				Content:   "content",
			},
			func() {
				s.updateRequestValid.EXPECT().
					Validate(&dto.UpdatePasteRequest{
						Content:  "content",
						Privacy:  string(domain.PublicPolicy),
						Password: "pass",
						ExpireAt: now.Add(2 * time.Hour),
					}).
					Return(nil).
					Once()

				s.repo.EXPECT().
					GetByHash(mock.Anything, "hash").
					Return(&domain.Paste{
						Id:        10,
						Views:     2,
						Privacy:   domain.PrivatePolicy,
						Hash:      "hash",
						CreatedAt: now,
						ExpireAt:  now.Add(2 * time.Hour),
						Content:   "content",
					}, nil).
					Once()

				s.repo.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(paste *domain.Paste) bool {
						return paste.Content == "content" &&
							paste.Privacy == domain.PublicPolicy &&
							paste.PasswordHash == "" && paste.ExpireAt.Equal(now.Add(2*time.Hour))
					})).
					Return(&domain.Paste{
						Id:        10,
						Views:     2,
						Privacy:   domain.PublicPolicy,
						Hash:      "hash",
						CreatedAt: now,
						ExpireAt:  now.Add(2 * time.Hour),
						Content:   "content",
					}, nil).Once()
			},
		},
		{
			"No updated fields with protected privacy",
			&dto.UpdatePasteRequest{
				Content:  "content",
				Privacy:  string(domain.ProtectedPolicy),
				ExpireAt: now.Add(time.Hour),
			},
			&dto.PasteResponse{
				Id:        10,
				Views:     2,
				Privacy:   string(domain.ProtectedPolicy),
				CreatedAt: now,
				ExpireAt:  now.Add(time.Hour),
				Content:   "content",
			},
			func() {
				s.updateRequestValid.EXPECT().
					Validate(&dto.UpdatePasteRequest{
						Content:  "content",
						Privacy:  string(domain.ProtectedPolicy),
						ExpireAt: now.Add(time.Hour),
					}).
					Return(nil).
					Once()

				s.repo.EXPECT().
					GetByHash(mock.Anything, "hash").
					Return(&domain.Paste{
						Id:           10,
						Views:        2,
						Privacy:      domain.PublicPolicy,
						Hash:         "hash",
						PasswordHash: "hash",
						CreatedAt:    now,
						ExpireAt:     now.Add(time.Hour),
						Content:      "content",
					}, nil).
					Once()

				s.repo.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(paste *domain.Paste) bool {
						return paste.Content == "content" &&
							paste.Privacy == domain.ProtectedPolicy &&
							paste.PasswordHash == "hash" && paste.ExpireAt.Equal(now.Add(time.Hour))
					})).
					Return(&domain.Paste{
						Id:        10,
						Views:     2,
						Privacy:   domain.ProtectedPolicy,
						Hash:      "hash",
						CreatedAt: now,
						ExpireAt:  now.Add(time.Hour),
						Content:   "content",
					}, nil).Once()
			},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			res, err := s.useCase.UpdatePaste(context.Background(), "hash", tc.value)

			s.Require().NoError(err)

			s.Equal(tc.expected.Id, res.Id)
			s.Equal(tc.expected.Views, res.Views)
			s.Equal(tc.expected.Privacy, res.Privacy)
			s.True(tc.expected.CreatedAt.Equal(res.CreatedAt))
			s.True(tc.expected.ExpireAt.Equal(res.ExpireAt))
			s.Equal(tc.expected.Content, res.Content)
		})
	}
}

func (s *UseCaseTestSuite) Test_UpdatePaste_Error() {
	cases := []struct {
		name          string
		value         *dto.UpdatePasteRequest
		expectedError error
		setup         func()
	}{
		{
			"Validation error",
			&dto.UpdatePasteRequest{},
			errs.ErrValidationProcess,
			func() {
				s.updateRequestValid.EXPECT().
					Validate(mock.Anything).
					Return(errs.ErrValidationProcess).
					Once()
			},
		},
		{
			"Get Paste By Hash - Not Found error",
			&dto.UpdatePasteRequest{},
			errs.ErrPasteNotFound,
			func() {
				s.updateRequestValid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.repo.EXPECT().
					GetByHash(mock.Anything, mock.Anything).
					Return(nil, errs.ErrPasteNotFound).
					Once()
			},
		},
		{
			"Get Paste By Hash - Internal error",
			&dto.UpdatePasteRequest{},
			errs.ErrInternal,
			func() {
				s.updateRequestValid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.repo.EXPECT().
					GetByHash(mock.Anything, mock.Anything).
					Return(nil, errors.New("some error")).
					Once()
			},
		},
		{
			"Hashing password error",
			&dto.UpdatePasteRequest{
				Privacy:  string(domain.ProtectedPolicy),
				Password: "pass",
			},
			errs.ErrInternal,
			func() {
				s.updateRequestValid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.repo.EXPECT().
					GetByHash(mock.Anything, mock.Anything).
					Return(&domain.Paste{}, nil).
					Once()

				s.security.EXPECT().
					CompareHashAndPassword(mock.Anything, mock.Anything).
					Return(false).
					Once()

				s.security.EXPECT().
					HashPassword(mock.Anything).
					Return("", errors.New("some error"))
			},
		},
		{
			"Update paste error",
			&dto.UpdatePasteRequest{},
			errs.ErrInternal,
			func() {
				s.updateRequestValid.EXPECT().
					Validate(mock.Anything).
					Return(nil).
					Once()

				s.repo.EXPECT().
					GetByHash(mock.Anything, mock.Anything).
					Return(&domain.Paste{}, nil).
					Once()

				s.repo.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(nil, errors.New("some error")).
					Once()
			},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			res, err := s.useCase.UpdatePaste(context.Background(), "hash", tc.value)

			s.Require().Nil(res)
			s.Require().Error(err)

			s.ErrorIs(err, tc.expectedError)
		})
	}
}
