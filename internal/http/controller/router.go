package controller

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func Router() http.Handler {
	r := chi.NewRouter()

	handler := NewHandler()

	r.Get("/", handler.HelloHandler)

	return r
}
