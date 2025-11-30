package main

import (
	"embed"
	"fmt"
	"hyperliquid-recon/api"
	"hyperliquid-recon/config"
	"hyperliquid-recon/services"
	"io/fs"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

//go:embed frontend/build
var frontendFS embed.FS

func main() {
	// Initialize reconciliation service
	reconService := services.NewReconciliationService()

	// Initialize API handler
	handler := api.NewHandler(reconService)

	// Setup router
	router := mux.NewRouter()

	// CORS middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// API routes
	router.HandleFunc("/api/health", handler.HealthCheck).Methods("GET")
	router.HandleFunc("/api/pnl", handler.GetPnLSummary).Methods("GET")
	router.HandleFunc("/api/refresh", handler.TriggerRefresh).Methods("POST")

	// Serve embedded frontend (production) or allow CORS for development
	if _, err := fs.Stat(frontendFS, "frontend/build/index.html"); err == nil {
		// Production mode: serve embedded frontend
		log.Println("Running in PRODUCTION mode (embedded frontend)")

		// Get the subdirectory
		buildFS, err := fs.Sub(frontendFS, "frontend/build")
		if err != nil {
			log.Fatal("Failed to get frontend build directory:", err)
		}

		// Serve static files
		staticServer := http.FileServer(http.FS(buildFS))

		// Handle SPA routing - serve index.html for all non-API routes
		router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to serve the file
			path := r.URL.Path
			if path == "/" {
				path = "/index.html"
			}

			// Check if file exists
			if _, err := fs.Stat(buildFS, path[1:]); err != nil {
				// File doesn't exist, serve index.html for SPA routing
				r.URL.Path = "/"
			}

			staticServer.ServeHTTP(w, r)
		})
	} else {
		// Development mode: CORS is already enabled above
		log.Println("Running in DEVELOPMENT mode (CORS enabled for external frontend)")
		log.Println("Note: Frontend should be running separately on port 3000")
	}

	// Start server
	addr := ":" + config.ServerPort
	fmt.Printf("Server starting on http://localhost%s\n", addr)
	if _, err := fs.Stat(frontendFS, "frontend/build/index.html"); err == nil {
		fmt.Printf("Access the application at: http://localhost%s\n", addr)
	}
	log.Fatal(http.ListenAndServe(addr, router))
}