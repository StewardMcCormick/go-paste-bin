package HTTP

import (
	"github.com/StewardMcCormick/Paste_Bin/internal/controller/HTTP/handlers"
	"github.com/StewardMcCormick/Paste_Bin/internal/controller/HTTP/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

func NewRouter(handler handlers.Handlers, logger *zap.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.RecovererMiddleware)

	r.Get("/hello", handler.HelloHandler)
	r.Get("/", handler.HelloHandler)

	return r
}
