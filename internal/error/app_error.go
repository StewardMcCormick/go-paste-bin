package error

import (
	"context"
	"encoding/json"
	"errors"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
	"net/http"
)

type AppError interface {
	error
	Code() int
}

func SendAppError(ctx context.Context, w http.ResponseWriter, status int, message error) {
	log := appctx.GetLogger(ctx)

	if !errors.Is(message, InternalError) {
		log.Info(message.Error())
	}

	var response AppError

	var validErr ValidationError
	if errors.As(message, &validErr) {
		validErr.Status = status
		response = validErr
	} else {
		response = BaseError{Message: message.Error(), Status: status}
	}

	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
