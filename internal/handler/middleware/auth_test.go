package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	"github.com/StewardMcCormick/Paste_Bin/internal/handler/middleware/mocks"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AuthMiddlewareTestSuite struct {
	suite.Suite
	auth       *mocks.MockAuthService
	middleware Auth
}

func TestAuthMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(AuthMiddlewareTestSuite))
}

func (s *AuthMiddlewareTestSuite) SetupTest() {
	s.auth = mocks.NewMockAuthService(s.T())

	s.middleware = NewAuth(s.auth)
}

func (s *AuthMiddlewareTestSuite) Test_Auth_Success() {
	validApiKey := "pb_test_abcd_acbdefghijkl"
	body := []byte(`{"message": "hello"`)

	var capturedContext context.Context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedContext = r.Context()
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	s.auth.EXPECT().
		Authenticate(mock.Anything, mock.Anything).
		Return(1, nil).
		Once()

	req := httptest.NewRequest("GET", "/", nil)

	req.Header.Set(APIKeyHeader, validApiKey)
	w := httptest.NewRecorder()

	midd := s.middleware.Handler(testHandler)

	midd.ServeHTTP(w, req)

	resultBody, err := io.ReadAll(w.Result().Body)
	s.Require().NoError(err)
	userIdFromCtx, err := appctx.GetUserId(capturedContext)
	s.Require().NoError(err)

	s.Equal(http.StatusOK, w.Result().StatusCode)
	s.Equal(string(body), string(resultBody))
	s.Equal(int64(1), userIdFromCtx)
}

func (s *AuthMiddlewareTestSuite) Test_Auth_Error() {
	cases := []struct {
		name       string
		setup      func()
		value      string
		wantStatus int
	}{
		{
			"Invalid API-Key",
			func() {
			},
			"invalid-key",
			http.StatusUnauthorized,
		},
		{
			"Auth Service Error - ErrUnauthorized",
			func() {
				s.auth.EXPECT().
					Authenticate(mock.Anything, mock.Anything).
					Return(0, errs.ErrUnauthorized).
					Once()
			},
			"pb_test_abcd_acbdefghijkl",
			http.StatusUnauthorized,
		},
		{
			"Auth Service Error - Internal Error",
			func() {
				s.auth.EXPECT().
					Authenticate(mock.Anything, mock.Anything).
					Return(0, errs.ErrInternal).
					Once()
			},
			"pb_test_abcd_acbdefghijkl",
			http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			callCount := 0
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				w.Write([]byte("hello"))
			})

			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set(APIKeyHeader, tc.value)
			w := httptest.NewRecorder()

			midd := s.middleware.Handler(testHandler)
			midd.ServeHTTP(w, req)

			s.Equal(tc.wantStatus, w.Result().StatusCode)
			s.Equal(0, callCount)
		})
	}
}
