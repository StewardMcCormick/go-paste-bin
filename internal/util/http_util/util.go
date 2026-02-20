package httputil

import (
	"net/http"
)

type WriterWithStatusCode struct {
	http.ResponseWriter
	StatusCode int
}

func (wc *WriterWithStatusCode) WriteHeader(status int) {
	wc.StatusCode = status
	wc.ResponseWriter.WriteHeader(status)
}
