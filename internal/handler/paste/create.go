package paste

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	"github.com/StewardMcCormick/Paste_Bin/pkg/render"
)

func (h *httpHandlers) Create(w http.ResponseWriter, r *http.Request) {
	req := &dto.PasteRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		errs.SendAppError(r.Context(), w, http.StatusBadRequest, fmt.Errorf("%w - invalid JSON", errs.BadRequest))
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

	pasteUrl := "/api/v1/paste/" + resp.Hash

	w.Header().Add("Location", pasteUrl)
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, resp)
}
