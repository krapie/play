package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

// httpRequests counts responses per router pattern and status class (revamp plan R7.1, block 3): the app's own view of
// errors, including calls that never pass the gateway (probes, in-cluster callers). The route is the pattern
// ("/api/urls/{code}"), never the raw path, so the series count stays at routes x 4 classes.
var httpRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "app_http_requests_total",
	Help: "HTTP responses by router pattern and status class.",
}, []string{"route", "code"})

func init() { prometheus.MustRegister(httpRequests) }

func statusClass(status int) string {
	switch {
	case status >= 500:
		return "5xx"
	case status >= 400:
		return "4xx"
	case status >= 300:
		return "3xx"
	}
	return "2xx"
}

// errorLog writes one line per 5xx response with the request id (the gateway's X-Request-Id, kept by
// middleware.RequestID). The collector keeps only app lines that contain an error word and chi's request
// line has none, so this line is what Loki holds next to the Envoy line with the same id (revamp plan R7.2).
// The id is client-supplied text: it finds one request and proves nothing about who sent it.
func errorLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		route := "unmatched"
		if rc := chi.RouteContext(r.Context()); rc != nil && rc.RoutePattern() != "" {
			route = rc.RoutePattern()
		}
		httpRequests.WithLabelValues(route, statusClass(ww.Status())).Inc()
		if ww.Status() >= 500 {
			log.Printf("error: %s %s -> %d [req %s]", r.Method, r.URL.Path, ww.Status(), chimw.GetReqID(r.Context()))
		}
	})
}
