package handler

import (
	mid "github.com/StewardMcCormick/Paste_Bin/internal/handler/middleware"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type UserHandler interface {
	Registration(w http.ResponseWriter, r *http.Request)
}

func NewRouter(
	userHandler UserHandler,
	logMid mid.Logging,
	recovererMid mid.Recoverer,
	envMid mid.Environmental,
) http.Handler {
	r := chi.NewRouter()

	r.Use(logMid.Handler)
	r.Use(recovererMid.Handler)
	r.Use(envMid.Handler)

	r.Post("/user", userHandler.Registration)

	return r
}
