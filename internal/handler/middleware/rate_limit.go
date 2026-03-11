package middleware

import (
	"context"
	"net/http"
	"strconv"

	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
)

type Limiter interface {
	AllowRequest(ctx context.Context, key string) (bool, error)
}

type IPLimiter struct {
	limiter Limiter
}

func NewIPLimiter(limiter Limiter) IPLimiter {
	return IPLimiter{limiter: limiter}
}

type UserIdLimiter struct {
	limiter Limiter
}

func NewUserIdLimiter(limiter Limiter) UserIdLimiter {
	return UserIdLimiter{limiter: limiter}
}

func (l *IPLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		allowed, err := l.limiter.AllowRequest(r.Context(), ip)
		if err != nil {
			errs.SendAppError(r.Context(), w, http.StatusInternalServerError, errs.InternalError)
			return
		}

		if !allowed {
			errs.SendAppError(r.Context(), w, http.StatusTooManyRequests, errs.TooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (l *UserIdLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := appctx.GetUserId(r.Context())
		if err != nil {
			errs.SendAppError(r.Context(), w, http.StatusInternalServerError, errs.InternalError)
			return
		}

		allowed, err := l.limiter.AllowRequest(r.Context(), strconv.FormatInt(id, 10))
		if err != nil {
			errs.SendAppError(r.Context(), w, http.StatusInternalServerError, err)
			return
		}

		if !allowed {
			errs.SendAppError(r.Context(), w, http.StatusTooManyRequests, errs.TooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
