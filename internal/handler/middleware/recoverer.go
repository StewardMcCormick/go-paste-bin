package middleware

import (
	"fmt"
	"net/http"

	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
	"go.uber.org/zap"
)

type Recoverer struct {
}

func NewRecoverer() Recoverer {
	return Recoverer{}
}

func (m *Recoverer) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger := r.Context().Value(appctx.LoggerKey).(*zap.Logger)
				logger.Error(fmt.Sprintf("[PANIC] %v", fmt.Sprint(err)))
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
