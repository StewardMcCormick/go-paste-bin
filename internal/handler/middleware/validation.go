package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
)

type JSONValidation struct {
}

func NewJSONValidation() JSONValidation {
	return JSONValidation{}
}

func (m *JSONValidation) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			errs.SendAppError(r.Context(), w, http.StatusBadRequest,
				fmt.Errorf("cannot read JSON request body: %v", body),
			)
			return
		}

		if len(body) == 0 {
			errs.SendAppError(r.Context(), w, http.StatusBadRequest, errors.New("request body is required"))
			return
		}

		if !json.Valid(body) {
			errs.SendAppError(r.Context(), w, http.StatusBadRequest,
				errors.New("invalid JSON"),
			)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))

		next.ServeHTTP(w, r)
	})
}
