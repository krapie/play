package main

import (
	"log"
	"net/http"

	chimw "github.com/go-chi/chi/v5/middleware"
)

// errorLog writes one line per 5xx response with the request id (the gateway's X-Request-Id, kept by
// middleware.RequestID). The collector keeps only app lines that contain an error word and chi's request
// line has none, so this line is what Loki holds next to the Envoy line with the same id (revamp plan R7.2).
// The id is client-supplied text: it finds one request and proves nothing about who sent it.
func errorLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		if ww.Status() >= 500 {
			log.Printf("error: %s %s -> %d [req %s]", r.Method, r.URL.Path, ww.Status(), chimw.GetReqID(r.Context()))
		}
	})
}
