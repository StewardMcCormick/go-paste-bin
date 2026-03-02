package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	"github.com/StewardMcCormick/Paste_Bin/pkg/render"
	"github.com/go-playground/validator/v10"
)

func (h *httpHandlers) Registration(w http.ResponseWriter, r *http.Request) {
	var userRequest dto.UserRequest
	json.NewDecoder(r.Body).Decode(&userRequest)

	user, err := h.authUseCase.Registration(r.Context(), &userRequest)
	if err != nil {
		if errors.Is(err, errs.UserAlreadyExists) {
			errs.SendAppError(r.Context(), w, http.StatusConflict, err)
			return
		} else if errors.As(err, &validator.ValidationErrors{}) {
			errs.SendAppError(r.Context(), w, http.StatusBadRequest, err)
			return
		}

		errs.SendAppError(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, user)
}
