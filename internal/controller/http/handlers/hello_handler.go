package handlers

import (
	"net/http"
)

func (h *Handler) HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"message": "Hello world!"}`))
}
