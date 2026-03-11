package middleware

import (
	"fmt"
	"net/http"

	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
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
				logger := appctx.GetLogger(r.Context())
				logger.Error(fmt.Sprintf("[PANIC] %v", fmt.Sprint(err)))
				errs.SendAppError(r.Context(), w, http.StatusInternalServerError, errs.InternalError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
