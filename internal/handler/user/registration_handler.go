package user

import (
	"encoding/json"
	"errors"
	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	"github.com/StewardMcCormick/Paste_Bin/pkg/render"
	"github.com/go-playground/validator/v10"
	"net/http"
)

func (h *httpHandlers) Registration(w http.ResponseWriter, r *http.Request) {
	var userRequest dto.CreateUserRequest
	json.NewDecoder(r.Body).Decode(&userRequest)

	user, err := h.UserUseCase.Registration(r.Context(), &userRequest)
	if err != nil {
		if errors.Is(err, errs.UserAlreadyExists) {
			errs.SendAppError(r.Context(), w, http.StatusConflict, err)
			return
		} else if errors.Is(err, errs.InternalError) {
			errs.SendAppError(r.Context(), w, http.StatusInternalServerError, err)
			return
		} else if errors.As(err, &validator.ValidationErrors{}) {
			errs.SendAppError(r.Context(), w, http.StatusBadRequest, err.(validator.ValidationErrors))
			return
		}
		errs.SendAppError(r.Context(), w, http.StatusBadRequest, err)
		return
	}

	render.JSON(w, user)
}
