package middleware

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StewardMcCormick/Paste_Bin/internal/handler/middleware/mocks"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type RateLimitMiddlewareTestSuite struct {
	suite.Suite
	ipLimiter               *mocks.MockLimiter
	userIdLimiter           *mocks.MockLimiter
	ipLimiterMiddleware     IPLimiter
	UserIdLimiterMiddleware UserIdLimiter
}

func TestRateLimitMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(RateLimitMiddlewareTestSuite))
}

func (s *RateLimitMiddlewareTestSuite) SetupTest() {
	s.ipLimiter = mocks.NewMockLimiter(s.T())
	s.userIdLimiter = mocks.NewMockLimiter(s.T())

	s.ipLimiterMiddleware = NewIPLimiter(s.ipLimiter)
	s.UserIdLimiterMiddleware = NewUserIdLimiter(s.userIdLimiter)
}

func (s *RateLimitMiddlewareTestSuite) Test_IPLimiter_AllowRequest() {
	callCount := 0
	body := []byte("Hello")
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	s.ipLimiter.EXPECT().
		AllowRequest(mock.Anything, mock.Anything).
		Return(true, nil).
		Once()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	midd := s.ipLimiterMiddleware.Handler(testHandler)

	midd.ServeHTTP(w, req)

	resultBody, err := io.ReadAll(w.Result().Body)
	s.Require().NoError(err)

	s.Equal(http.StatusOK, w.Result().StatusCode)
	s.Equal(string(body), string(resultBody))
	s.Equal(1, callCount)
}

func (s *RateLimitMiddlewareTestSuite) Test_IPLimiter_Error() {
	cases := []struct {
		name         string
		setup        func()
		expectedCode int
	}{
		{"Internal error",
			func() {
				s.ipLimiter.EXPECT().
					AllowRequest(mock.Anything, mock.Anything).
					Return(false, errors.New("some internal error")).
					Once()
			},
			http.StatusInternalServerError,
		},
		{
			"Too Many Requests Error",
			func() {
				s.ipLimiter.EXPECT().
					AllowRequest(mock.Anything, mock.Anything).
					Return(false, nil).
					Once()
			},
			http.StatusTooManyRequests,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			callCount := 0
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			midd := s.ipLimiterMiddleware.Handler(testHandler)

			midd.ServeHTTP(w, req)

			s.Equal(tc.expectedCode, w.Result().StatusCode)
			s.Equal(0, callCount)
		})
	}
}

func (s *RateLimitMiddlewareTestSuite) Test_UserIdLimiter_AllowRequest() {
	callCount := 0
	body := []byte("Hello")
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	s.userIdLimiter.EXPECT().
		AllowRequest(mock.Anything, mock.Anything).
		Return(true, nil).
		Once()

	req := httptest.NewRequest("GET", "/", nil)
	ctx := appctx.WithUserId(req.Context(), 1)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	midd := s.UserIdLimiterMiddleware.Handler(testHandler)

	midd.ServeHTTP(w, req)

	resultBody, err := io.ReadAll(w.Result().Body)
	s.Require().NoError(err)

	s.Equal(http.StatusOK, w.Result().StatusCode)
	s.Equal(string(body), string(resultBody))
	s.Equal(1, callCount)
}

func (s *RateLimitMiddlewareTestSuite) Test_UserIdLimiter_WithIncorrectUserId() {
	callCount := 0
	body := []byte("Hello")
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), appctx.UserIdKey, "incorrect id")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	midd := s.UserIdLimiterMiddleware.Handler(testHandler)

	midd.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Result().StatusCode)
	s.Equal(0, callCount)
}

func (s *RateLimitMiddlewareTestSuite) Test_UserIdLimiter_Error() {
	cases := []struct {
		name         string
		setup        func()
		expectedCode int
	}{
		{"Internal error",
			func() {
				s.userIdLimiter.EXPECT().
					AllowRequest(mock.Anything, "0").
					Return(false, errors.New("some internal error")).
					Once()
			},
			http.StatusInternalServerError,
		},
		{
			"Too Many Requests Error",
			func() {
				s.userIdLimiter.EXPECT().
					AllowRequest(mock.Anything, "0").
					Return(false, nil).
					Once()
			},
			http.StatusTooManyRequests,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			callCount := 0
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/", nil)
			ctx := appctx.WithUserId(req.Context(), 0)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			midd := s.UserIdLimiterMiddleware.Handler(testHandler)

			midd.ServeHTTP(w, req)

			s.Equal(tc.expectedCode, w.Result().StatusCode)
			s.Equal(0, callCount)
		})
	}
}
