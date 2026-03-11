package paste

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	"github.com/StewardMcCormick/Paste_Bin/pkg/render"
	"github.com/go-chi/chi/v5"
)

func (h *httpHandlers) GetPaste(w http.ResponseWriter, r *http.Request) {
	pasteHash := chi.URLParam(r, "pasteHash")

	req := dto.GetPasteRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		errs.SendAppError(r.Context(), w, http.StatusBadRequest, fmt.Errorf("%w - invalid JSON", errs.BadRequest))
		return
	}

	result, err := h.useCase.GetByHash(r.Context(), req, pasteHash)
	if err != nil {
		if errors.Is(err, errs.PasteNotFound) {
			errs.SendAppError(r.Context(), w, http.StatusNotFound, err)
			return
		} else if errors.Is(err, errs.Forbidden) {
			errs.SendAppError(r.Context(), w, http.StatusForbidden, err)
			return
		} else if errors.Is(err, errs.Unauthorized) {
			errs.SendAppError(r.Context(), w, http.StatusUnauthorized, err)
			return
		}

		errs.SendAppError(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	render.JSON(w, result)
}
