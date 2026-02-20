package error

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/StewardMcCormick/Paste_Bin/internal/util"
	"net/http"
)

type AppError interface {
	error
	Code() int
}

func SendAppError(ctx context.Context, w http.ResponseWriter, status int, message error) {
	log := util.GetLoggerFromCtx(ctx)

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
	log.Info(message.Error())
}
