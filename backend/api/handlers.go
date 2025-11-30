package api

import (
	"encoding/json"
	"hyperliquid-recon/config"
	"hyperliquid-recon/services"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Handler handles HTTP requests for the reconciliation API
type Handler struct {
	reconService *services.ReconciliationService
}

// Response represents a standard API response
type Response struct {
	Status  string      `json:"status,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// NewHandler creates a new API handler
func NewHandler(reconService *services.ReconciliationService) *Handler {
	return &Handler{
		reconService: reconService,
	}
}

// respondWithJSON writes a JSON response
func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// respondWithError writes a JSON error response
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	respondWithJSON(w, statusCode, ErrorResponse{Error: message})
}

// GetPnLSummary handles GET /api/pnl requests
func (h *Handler) GetPnLSummary(w http.ResponseWriter, r *http.Request) {
	summary := h.reconService.GetPnLSummary()
	respondWithJSON(w, http.StatusOK, summary)
}

// TriggerRefresh handles POST /api/refresh requests
func (h *Handler) TriggerRefresh(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		respondWithError(w, http.StatusBadRequest, "address parameter is required")
		return
	}

	// Parse days parameter, default to config value if not provided
	days := config.TradeHistoryDays
	daysParam := r.URL.Query().Get("days")
	if daysParam != "" {
		parsedDays, err := strconv.Atoi(daysParam)
		if err != nil || parsedDays <= 0 {
			respondWithError(w, http.StatusBadRequest, "days parameter must be a positive integer")
			return
		}
		days = parsedDays
	}

	if err := h.reconService.FetchAndReconcile(address, days); err != nil {
		log.Printf("Error fetching and reconciling trades for %s (days=%d): %v", address, days, err)

		// Provide more specific error messages
		errorMsg := err.Error()
		statusCode := http.StatusInternalServerError

		// Check if it's a rate limiting error
		if contains(errorMsg, "429") || contains(errorMsg, "rate limit") {
			errorMsg = "Rate limit exceeded. Please wait a moment before refreshing again."
			statusCode = http.StatusTooManyRequests
		} else if contains(errorMsg, "timeout") {
			errorMsg = "Request timeout. Please try again."
			statusCode = http.StatusGatewayTimeout
		} else {
			errorMsg = "Failed to refresh data. Please try again later."
		}

		respondWithError(w, statusCode, errorMsg)
		return
	}

	respondWithJSON(w, http.StatusOK, Response{
		Status:  "success",
		Message: "Data refreshed successfully",
	})
}

// HealthCheck handles GET /api/health requests
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, Response{
		Status:  "healthy",
		Message: "Service is running",
	})
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
