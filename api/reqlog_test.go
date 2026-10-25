package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/testutil"
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

func TestRequestCounterUsesRoutePattern(t *testing.T) {
	httpRequests.Reset()
	r := chi.NewRouter()
	r.Use(errorLog)
	r.Get("/item/{id}", func(w http.ResponseWriter, r *http.Request) {})
	r.Get("/fail", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusBadGateway) })
	for _, p := range []string{"/item/1", "/item/2", "/fail", "/nope"} {
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", p, nil))
	}
	if n := testutil.ToFloat64(httpRequests.WithLabelValues("/item/{id}", "2xx")); n != 2 {
		t.Fatalf("/item/{id} 2xx = %v, want 2 (the pattern, not the raw path)", n)
	}
	if n := testutil.ToFloat64(httpRequests.WithLabelValues("/fail", "5xx")); n != 1 {
		t.Fatalf("/fail 5xx = %v, want 1", n)
	}
	if n := testutil.ToFloat64(httpRequests.WithLabelValues("unmatched", "4xx")); n != 1 {
		t.Fatalf("unmatched 4xx = %v, want 1", n)
	}
}
