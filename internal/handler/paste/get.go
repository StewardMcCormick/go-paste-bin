package paste

import (
	"errors"
	"net/http"

	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	"github.com/StewardMcCormick/Paste_Bin/pkg/render"
	"github.com/go-chi/chi/v5"
)

func (h *httpHandlers) GetPaste(w http.ResponseWriter, r *http.Request) {
	pasteHash := chi.URLParam(r, "pasteHash")

	result, err := h.useCase.GetByHash(r.Context(), pasteHash)
	if err != nil {
		if errors.Is(err, errs.PasteNotFound) {
			errs.SendAppError(r.Context(), w, http.StatusNotFound, err)
			return
		} else if errors.Is(err, errs.Forbidden) {
			errs.SendAppError(r.Context(), w, http.StatusForbidden, err)
			return
		}

		errs.SendAppError(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	render.JSON(w, result)
}
