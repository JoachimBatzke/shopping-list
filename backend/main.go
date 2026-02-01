package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// Rate limiter - simple in-memory implementation
var (
	rateLimitMu sync.Mutex
	rateLimits  = make(map[string]*rateLimitEntry)
)

type rateLimitEntry struct {
	count     int
	resetTime time.Time
}

const (
	rateLimit       = 100 // requests per window
	rateLimitWindow = time.Minute
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

	// Item routes (nested under lists for security - verifies list ownership)
	mux.HandleFunc("GET /api/lists/{listId}/items", GetItems)
	mux.HandleFunc("POST /api/lists/{listId}/items", CreateItem)
	mux.HandleFunc("PUT /api/lists/{listId}/items/reorder", ReorderItems)
	mux.HandleFunc("PATCH /api/lists/{listId}/items/{id}", UpdateItem)
	mux.HandleFunc("DELETE /api/lists/{listId}/items/{id}", DeleteItem)

	// Recommendations routes
	mux.HandleFunc("GET /api/lists/{listId}/recommendations", GetRecommendations)
	mux.HandleFunc("POST /api/lists/{listId}/recommendations/{name}/dismiss", DismissRecommendation)

	// Health check endpoint (useful for deployment platforms)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Wrap with middleware chain: rate limiting -> CORS -> router
	handler := rateLimitMiddleware(corsMiddleware(mux))

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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests (browsers send OPTIONS before actual request)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// rateLimitMiddleware prevents brute-force attacks by limiting requests per IP
func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for health checks
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Get client IP (check X-Forwarded-For for proxied requests)
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = strings.Split(forwarded, ",")[0]
		}

		rateLimitMu.Lock()
		entry, exists := rateLimits[ip]
		now := time.Now()

		if !exists || now.After(entry.resetTime) {
			// New window
			rateLimits[ip] = &rateLimitEntry{
				count:     1,
				resetTime: now.Add(rateLimitWindow),
			}
			rateLimitMu.Unlock()
			next.ServeHTTP(w, r)
			return
		}

		if entry.count >= rateLimit {
			rateLimitMu.Unlock()
			http.Error(w, "Rate limit exceeded. Try again later.", http.StatusTooManyRequests)
			return
		}

		entry.count++
		rateLimitMu.Unlock()
		next.ServeHTTP(w, r)
	})
}
