package config

import "time"

const (
	// ServerPort Server configuration
	ServerPort = "8080"

	// HyperliquidAPIURL Hyperliquid API configuration
	HyperliquidAPIURL = "https://api.hyperliquid.xyz/info"
	APITimeout        = 30 * time.Second

	// TradeHistoryDays Data fetching configuration
	TradeHistoryDays  = 10
	MaxTradesPerBatch = 2000
	RateLimitDelayMs  = 300
	RateLimitDelay    = RateLimitDelayMs * time.Millisecond
)
