package middleware

import (
	"context"
	"github.com/StewardMcCormick/Paste_Bin/config/cfg_util"
	httpUtil "github.com/StewardMcCormick/Paste_Bin/internal/util/http_util"
	"net/http"
)

type Environmental struct {
	env cfgUtil.Env
}

func NewEnv(env cfgUtil.Env) Environmental {
	return Environmental{env}
}

func (m *Environmental) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.env == "" || (m.env != cfgUtil.ProductionEnv && m.env != cfgUtil.DevelopmentEnv) {
			m.env = cfgUtil.ProductionEnv
		}

		ctx := context.WithValue(r.Context(), httpUtil.EnvKey, m.env)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
