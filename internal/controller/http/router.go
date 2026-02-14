package http

import (
	"github.com/StewardMcCormick/Paste_Bin/internal/controller/http/handlers"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

func Router(logger *zap.Logger) http.Handler {
	r := chi.NewRouter()

	h := handlers.NewHandler()

	r.Use(LoggerMiddleware(logger))

	r.Get("/", h.HelloHandler)

	return r
}
