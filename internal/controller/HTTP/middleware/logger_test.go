package middleware

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"net/http"
	"net/http/httptest"
	"testing"
)

type LoggerMiddlewareTestSuite struct {
	suite.Suite
	logger *zap.Logger
}

func TestLoggerMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(LoggerMiddlewareTestSuite))
}

func (s *LoggerMiddlewareTestSuite) SetupTest() {
	s.logger = zaptest.NewLogger(s.T())
}

func (s *LoggerMiddlewareTestSuite) Test_CreatingRequestId() {
	var requestIdFromMiddleware string

	handler := LoggerMiddleware(s.logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestIdFromMiddleware = r.Context().Value(requestIdKey).(string)
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)
	s.NotEmpty(requestIdFromMiddleware)
	_, err := uuid.Parse(requestIdFromMiddleware)
	s.NoError(err)
}

func (s *LoggerMiddlewareTestSuite) Test_WithExistingRequestId() {
	var requestIdFromMiddleware string

	handler := LoggerMiddleware(s.logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestIdFromMiddleware = r.Context().Value(requestIdKey).(string)
		w.WriteHeader(http.StatusOK)
	}))

	testRequestId := uuid.NewString()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Request-ID", testRequestId)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	s.Equal(testRequestId, requestIdFromMiddleware)
	_, err := uuid.Parse(requestIdFromMiddleware)
	s.NoError(err)
}

//func (s *RecovererMiddlewareTestSuite) Test_
