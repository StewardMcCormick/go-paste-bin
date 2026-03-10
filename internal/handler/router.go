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
}

type PasteHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetPaste(w http.ResponseWriter, r *http.Request)
}

func NewRouter(
	userHandler UserHandler,
	pasteHandler PasteHandler,
	logMid mid.Logging,
	ipRateLimitMid mid.IPLimiter,
	recovererMid mid.Recoverer,
	envMid mid.Environmental,
	validMid mid.JSONValidation,
	authMid mid.Auth,
	userIdRateLimitMid mid.UserIdLimiter,
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

	r.Group(func(r chi.Router) {
		r.Use(validMid.Handler)
		r.Use(ipRateLimitMid.Handler)

		r.Post("/registration", userHandler.Registration)
		r.Post("/login", userHandler.Login)
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(validMid.Handler)
		r.Use(authMid.Handler)
		r.Use(userIdRateLimitMid.Handler)

		r.Route("/paste", func(r chi.Router) {

			r.Group(func(r chi.Router) {
				r.Use(validMid.Handler)
				r.Post("/", pasteHandler.Create)
				r.Get("/{pasteHash}", pasteHandler.GetPaste)
			})

		})
	})

	return r
}
