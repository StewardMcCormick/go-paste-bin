package middleware

import (
	"net/http"

	"github.com/StewardMcCormick/Paste_Bin/config/cfg_util"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
)

type Environmental struct {
	env cfgutil.Env
}

func NewEnv(env cfgutil.Env) Environmental {
	return Environmental{env}
}

func (m *Environmental) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.env == "" || (m.env != cfgutil.ProductionEnv && m.env != cfgutil.DevelopmentEnv) {
			m.env = cfgutil.ProductionEnv
		}

		ctx := appctx.WithEnv(r.Context(), m.env)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
