package HTTP

import (
	"github.com/StewardMcCormick/Paste_Bin/internal/controller/HTTP/handlers/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zaptest"
	"net/http"
	"net/http/httptest"
	"testing"
)

type RouterTestSuite struct {
	suite.Suite
	handler *mocks.MockHandler
	router  http.Handler
}

func TestMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}

func (s *RouterTestSuite) SetupTest() {
	s.handler = mocks.NewMockHandler(s.T())
	s.router = NewRouter(s.handler, zaptest.NewLogger(s.T()))
}

func (s *RouterTestSuite) TestRouterRecovererMiddleware_OnPanic() {
	s.handler.EXPECT().
		HelloHandler(mock.Anything, mock.Anything).
		Run(func(w http.ResponseWriter, r *http.Request) {
			panic("foo")
		}).Once()
	r := httptest.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, r)

	s.Equal(http.StatusInternalServerError, w.Result().StatusCode)
}
