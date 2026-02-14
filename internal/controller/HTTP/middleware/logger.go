package middleware

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func LoggerMiddleware(logger *zap.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			requestId := r.Header.Get("X-Request-ID")
			if requestId == "" {
				requestId = uuid.NewString()
			}
			wrapped := &writerWithStatusCode{w, http.StatusOK}

			reqLogger := logger.With(
				zap.String("request_id", requestId),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("host", r.Host),
				zap.String("user_agent", r.UserAgent()),
			)

			reqLogger.Info("[NEW REQUEST]")

			ctx := context.WithValue(r.Context(), loggerKey, reqLogger)
			ctx = context.WithValue(ctx, requestIdKey, requestId)

			next.ServeHTTP(wrapped, r.WithContext(ctx))

			duration := time.Since(start)

			reqLogger.Info(
				"[REQUEST COMPLETED]",
				zap.Int64("duration_millis", duration.Milliseconds()),
				zap.Int("status_code", wrapped.statusCode),
			)
		})
	}
}
