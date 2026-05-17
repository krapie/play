package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/krapie/play-api/db"
	"github.com/krapie/play-api/handlers"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/data/play.db"
	}

	database, err := db.Connect(dbPath)
	if err != nil {
		log.Fatalf("db.Connect: %v", err)
	}
	defer database.Close()

	h := handlers.New(database)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	r.Route("/api/results", func(r chi.Router) {
		r.Post("/", h.Submit)
		r.Get("/{game}/leaderboard", h.Leaderboard)
		r.Get("/{game}/player/{pid}", h.PlayerStats)
	})

	log.Println("play-api listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "https://play.kevinprk.com")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
