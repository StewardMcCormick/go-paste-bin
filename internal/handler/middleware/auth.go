package middleware

import (
	"context"
	"errors"
	"net/http"

	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
	"github.com/StewardMcCormick/Paste_Bin/internal/util/validation"
)

const (
	APIKeyHeader = "X-API-Key"
)

type AuthService interface {
	Authenticate(ctx context.Context, apiKey string) (userId int64, err error)
}

type Auth struct {
	auth AuthService
}

func NewAuth(auth AuthService) Auth {
	return Auth{auth: auth}
}

func (a *Auth) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := appctx.GetLogger(r.Context())
		key := r.Header.Get(APIKeyHeader)
		if !validation.ValidateAPIKey(key) {
			log.Info("invalid key")
			errs.SendAppError(r.Context(), w, http.StatusUnauthorized, errs.ErrUnauthorized)
			return
		}

		userId, err := a.auth.Authenticate(r.Context(), key)
		if err != nil {
			if errors.Is(err, errs.ErrUnauthorized) {
				errs.SendAppError(r.Context(), w, http.StatusUnauthorized, err)
				return
			}

			errs.SendAppError(r.Context(), w, http.StatusInternalServerError, errs.ErrInternal)
			return
		}

		ctx := appctx.WithUserId(r.Context(), userId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
