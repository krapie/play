package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	chimw "github.com/go-chi/chi/v5/middleware"
)

func TestErrorLogCarriesRequestID(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	h := chimw.RequestID(errorLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/boom" {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})))
	for _, path := range []string{"/ok", "/boom"} {
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("X-Request-Id", "abc-123")
		h.ServeHTTP(httptest.NewRecorder(), req)
	}
	out := buf.String()
	if strings.Count(out, "error:") != 1 || !strings.Contains(out, "GET /boom -> 500 [req abc-123]") {
		t.Fatalf("want exactly one error line for /boom with the id, got %q", out)
	}
}
