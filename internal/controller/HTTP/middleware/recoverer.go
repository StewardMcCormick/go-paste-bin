package middleware

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
)

func RecovererMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger := r.Context().Value(loggerKey).(*zap.Logger)
				logger.Error(fmt.Sprintf("[ERROR] %v", fmt.Sprint(err)))
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
