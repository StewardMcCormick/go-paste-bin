package paste

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	"github.com/StewardMcCormick/Paste_Bin/pkg/render"
)

// TODO - add Location Header

func (h *httpHandlers) Create(w http.ResponseWriter, r *http.Request) {
	req := &dto.PasteRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		errs.SendAppError(r.Context(), w, http.StatusBadRequest, err)
		return
	}

	resp, err := h.useCase.Create(r.Context(), req)
	if err != nil {
		if errors.Is(err, errs.InternalError) {
			errs.SendAppError(r.Context(), w, http.StatusInternalServerError, err)
			return
		}
		errs.SendAppError(r.Context(), w, http.StatusBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, resp)
}
