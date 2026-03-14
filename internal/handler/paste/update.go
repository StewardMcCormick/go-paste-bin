package paste

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	"github.com/StewardMcCormick/Paste_Bin/pkg/render"
	"github.com/go-chi/chi/v5"
)

func (h *httpHandlers) UpdatePaste(w http.ResponseWriter, r *http.Request) {
	req := &dto.UpdatePasteRequest{}
	hash := chi.URLParam(r, "pasteHash")

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		errs.SendAppError(r.Context(), w, http.StatusBadRequest, fmt.Errorf("%w - invalid JSON", errs.BadRequest))
		return
	}

	resp, err := h.useCase.UpdatePaste(r.Context(), hash, req)
	if err != nil {
		if errors.Is(err, errs.PasteNotFound) {
			errs.SendAppError(r.Context(), w, http.StatusNotFound, err)
			return
		} else if errors.Is(err, errs.InternalError) {
			errs.SendAppError(r.Context(), w, http.StatusInternalServerError, err)
			return
		}
		errs.SendAppError(r.Context(), w, http.StatusBadRequest, err)
		return
	}

	render.JSON(w, resp)
}
