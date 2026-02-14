package middleware

import (
	"net/http"
)

type loggerCtxKey string
type requestIdCtxKey string

var (
	loggerKey    loggerCtxKey    = "logger"
	requestIdKey requestIdCtxKey = "request_id"
)

type writerWithStatusCode struct {
	http.ResponseWriter
	statusCode int
}

func (wc *writerWithStatusCode) WriteHeader(status int) {
	wc.statusCode = status
	wc.ResponseWriter.WriteHeader(status)
}
