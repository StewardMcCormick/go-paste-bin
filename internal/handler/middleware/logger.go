package middleware

import (
	"net/http"
	"time"

	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
	httputil "github.com/StewardMcCormick/Paste_Bin/internal/util/http_util"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Logging struct {
	logger *zap.Logger
}

func NewLogging(logger *zap.Logger) Logging {
	return Logging{logger: logger}
}

func (m *Logging) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestId := r.Header.Get("X-Request-ID")
		if requestId == "" {
			requestId = uuid.NewString()
		}
		wrapped := &httputil.WriterWithStatusCode{ResponseWriter: w, StatusCode: http.StatusOK}

		reqLogger := m.logger.With(
			zap.String("request_id", requestId),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("host", r.Host),
			zap.String("user_agent", r.UserAgent()),
		)

		reqLogger.Info("[NEW REQUEST]")

		ctx := appctx.WithLogger(r.Context(), reqLogger)
		ctx = appctx.WithRequestId(ctx, requestId)

		next.ServeHTTP(wrapped, r.WithContext(ctx))

		duration := time.Since(start)

		reqLogger.Info(
			"[REQUEST COMPLETED]",
			zap.Int64("duration_millis", duration.Milliseconds()),
			zap.Int("status_code", wrapped.StatusCode),
		)
	})
}
