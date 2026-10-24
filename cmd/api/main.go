package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"zy-library-manager/internal/database"

	"github.com/joho/godotenv"
)

func init() {
	/// load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("ENV files not found")
	}
}

func main() {

	// load DATABASE_URL
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	// test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	db, err := database.Open(ctx, dsn)
	cancel()
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	// endpoint /health
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		}); err != nil {
			log.Printf("write health response: %v", err)
		}
	})

	// endpoint /ready
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)

			if err := json.NewEncoder(w).Encode(map[string]string{
				"status": "not ready",
			}); err != nil {
				log.Printf("write readiness response: %v", err)
			}

			return
		}

		if err := json.NewEncoder(w).Encode(map[string]string{
			"status": "ready",
		}); err != nil {
			log.Printf("write readiness response: %v", err)
		}
	})

	addr := ":8080"

	log.Printf("API listening on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("API server stopped: %v", err)
	}

}
