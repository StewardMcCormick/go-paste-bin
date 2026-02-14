package handlers

import (
	http2 "github.com/StewardMcCormick/Paste_Bin/internal/controller/http"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	router = http2.Router(zap.L())
)

func TestHandler_HelloHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	body, _ := io.ReadAll(w.Body)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	assert.Equal(t, `{"message": "Hello world!"}`, string(body))
}
