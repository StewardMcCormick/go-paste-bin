package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestLogging_Handler(t *testing.T) {
	logger := zap.L()

	callCount := 0
	var capturedContext context.Context
	body := []byte("Hello")
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		capturedContext = r.Context()
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler := NewLogging(logger)
	midd := handler.Handler(testHandler)

	midd.ServeHTTP(w, req)

	resultLogger := appctx.GetLogger(capturedContext)
	resultBody, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)

	assert.Equal(t, 1, callCount)
	assert.Equal(t, logger, resultLogger)
	assert.Equal(t, string(resultBody), string(body))
}
