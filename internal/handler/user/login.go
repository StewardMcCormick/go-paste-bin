package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	"github.com/StewardMcCormick/Paste_Bin/pkg/render"
)

func (h *httpHandlers) Login(w http.ResponseWriter, r *http.Request) {
	user := &dto.UserRequest{}

	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		errs.SendAppError(r.Context(), w, http.StatusBadRequest, err)
		return
	}

	key, err := h.authUseCase.Login(r.Context(), user)
	if err != nil {
		if errors.Is(err, errs.UserNotFound) {
			errs.SendAppError(r.Context(), w, http.StatusNotFound, err)
			return
		} else if errors.Is(err, errs.Unauthorized) {
			errs.SendAppError(r.Context(), w, http.StatusUnauthorized, err)
			return
		} else if errors.As(err, &errs.ValidationError{}) {
			errs.SendAppError(r.Context(), w, http.StatusBadRequest, err)
			return
		}

		errs.SendAppError(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	render.JSON(w, key)
}
