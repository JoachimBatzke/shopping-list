package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file (only needed for local development)
	// In production (Railway), environment variables are set in the dashboard
	godotenv.Load()

	// Connect to database
	ConnectDB()
	defer CloseDB() // This runs when main() exits

	// Create a new router (Go 1.22+ has built-in routing with path parameters)
	mux := http.NewServeMux()

	// List routes
	// mux.HandleFunc("GET /api/lists", GetLists) // Disabled for privacy - lists only accessible by ID
	mux.HandleFunc("POST /api/lists", CreateList)
	mux.HandleFunc("GET /api/lists/{id}", GetList)
	mux.HandleFunc("PATCH /api/lists/{id}", UpdateList)
	mux.HandleFunc("DELETE /api/lists/{id}", DeleteList)

	// Item routes (nested under lists)
	mux.HandleFunc("GET /api/lists/{listId}/items", GetItems)
	mux.HandleFunc("POST /api/lists/{listId}/items", CreateItem)
	mux.HandleFunc("PATCH /api/items/{id}", UpdateItem)
	mux.HandleFunc("DELETE /api/items/{id}", DeleteItem)

	// Health check endpoint (useful for deployment platforms)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Wrap with CORS middleware
	handler := corsMiddleware(mux)

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

// corsMiddleware adds CORS headers to allow frontend to call the API
// Without this, browsers block requests from different origins (ports count as different!)
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get allowed origin from environment, default to * for development
		allowedOrigin := os.Getenv("CORS_ORIGIN")
		if allowedOrigin == "" {
			allowedOrigin = "*"
		}

		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests (browsers send OPTIONS before actual request)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
