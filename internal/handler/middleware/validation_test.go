package middleware

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type brokenBody struct {
}

func (b *brokenBody) Read(p []byte) (n int, err error) {
	return 0, errors.New("broken reader")
}

func TestJSONValidation_Handler(t *testing.T) {
	cases := []struct {
		name             string
		value            io.Reader
		method           string
		expectedCode     int
		expectedNextCall int
	}{
		{
			"On GET method",
			nil,
			"GET",
			http.StatusOK,
			1,
		},
		{
			"Valid JSON",
			bytes.NewReader([]byte(`{"message": "hello"}`)),
			"POST",
			http.StatusOK,
			1,
		},
		{
			"With broken body",
			&brokenBody{},
			"POST",
			http.StatusBadRequest,
			0,
		},
		{
			"Empty JSON",
			bytes.NewReader([]byte("")),
			"POST",
			http.StatusBadRequest,
			0,
		},
		{
			"Invalid JSON",
			bytes.NewReader([]byte(`{"message": ,}`)),
			"POST",
			http.StatusBadRequest,
			0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			callCount := 0
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(tc.method, "/", tc.value)
			w := httptest.NewRecorder()

			handler := NewJSONValidation()
			midd := handler.Handler(testHandler)

			midd.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Result().StatusCode)
			assert.Equal(t, tc.expectedNextCall, callCount)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		})
	}
}
