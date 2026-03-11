package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoverer_Handler_OnPanic(t *testing.T) {
	callCount := 0
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler := NewRecoverer()
	midd := handler.Handler(testHandler)

	midd.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
	assert.Equal(t, 1, callCount)
}

func TestRecoverer_Handler_WithoutPanic(t *testing.T) {
	callCount := 0
	body := []byte("Hello")
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler := NewRecoverer()
	midd := handler.Handler(testHandler)

	midd.ServeHTTP(w, req)

	resultBody, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	assert.Equal(t, 1, callCount)
	assert.Equal(t, string(body), string(resultBody))
}
