package handler

import (
	"net/http"

	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	mid "github.com/StewardMcCormick/Paste_Bin/internal/handler/middleware"
	"github.com/go-chi/chi/v5"
)

type UserHandler interface {
	Registration(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Hello(w http.ResponseWriter, r *http.Request)
}

type PasteHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
}

func NewRouter(
	userHandler UserHandler,
	pasteHandler PasteHandler,
	logMid mid.Logging,
	recovererMid mid.Recoverer,
	envMid mid.Environmental,
	validMid mid.JSONValidation,
	authMid mid.Auth,
) http.Handler {
	r := chi.NewRouter()
	
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		errs.SendAppError(r.Context(), w, http.StatusNotFound, errs.PageNotFound)
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		errs.SendAppError(r.Context(), w, http.StatusMethodNotAllowed, errs.MethodNotAllowed)
	})

	r.Use(logMid.Handler)
	r.Use(recovererMid.Handler)
	r.Use(envMid.Handler)
	r.Use(validMid.Handler)

	r.Post("/registration", userHandler.Registration)
	r.Post("/login", userHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(authMid.Handler)
		r.Get("/hello", userHandler.Hello)

		r.Route("/api/v1/paste", func(r chi.Router) {
			r.Post("/", pasteHandler.Create)
		})
	})

	return r
}
