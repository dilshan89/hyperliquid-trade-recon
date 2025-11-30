package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hyperliquid-recon/config"
	"hyperliquid-recon/models"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

// HyperliquidClient Client for interacting with the Hyperliquid API
type HyperliquidClient struct {
	httpClient *http.Client
}

func NewHyperliquidClient() *HyperliquidClient {
	return &HyperliquidClient{
		httpClient: &http.Client{Timeout: config.APITimeout},
	}
}

// UserFillsRequest represents the request body for fetching user fills
type UserFillsRequest struct {
	Type            string `json:"type"`
	User            string `json:"user"`
	StartTime       *int64 `json:"startTime,omitempty"`
	EndTime         *int64 `json:"endTime,omitempty"`
	AggregateByTime bool   `json:"aggregateByTime,omitempty"`
}

// FillResponse represents a single fill from the Hyperliquid API
type FillResponse struct {
	Time          int64  `json:"time"`
	Coin          string `json:"coin"`
	Side          string `json:"side"`
	Price         string `json:"px"`
	Size          string `json:"sz"`
	StartPosition string `json:"startPosition"`
	Dir           string `json:"dir"`
	ClosedPnl     string `json:"closedPnl"`
}

// FetchTrades fetches historical trades for a given address from Hyperliquid API
// It handles pagination automatically and returns all trades within the specified history period
func (c *HyperliquidClient) FetchTrades(address string, days int) ([]models.Trade, error) {
	// Calculate start time based on specified history days
	now := time.Now()
	historyStart := now.Add(-time.Duration(days) * 24 * time.Hour)

	return c.FetchTradesInRange(address, historyStart, now)
}

// FetchTradesInRange fetches trades for a given address within a specific time range
func (c *HyperliquidClient) FetchTradesInRange(address string, start, end time.Time) ([]models.Trade, error) {
	startTime := start.UnixMilli()
	endTime := end.UnixMilli()

	log.Printf("Fetching trades for %s from %s to %s", address, start.Format(time.RFC3339), end.Format(time.RFC3339))

	allTrades := make([]models.Trade, 0)
	currentStartTime := startTime

	// Pagination loop: fetch in batches
	batchCount := 0
	for {
		// Add delay between requests to avoid rate limiting (except first request)
		if batchCount > 0 {
			time.Sleep(config.RateLimitDelay)
		}
		batchCount++

		fills, err := c.fetchBatch(address, currentStartTime, endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch batch %d: %w", batchCount, err)
		}

		// If no more fills, break
		if len(fills) == 0 {
			log.Printf("Fetched %d trades in %d batches", len(allTrades), batchCount)
			break
		}

		// Convert fills to trades
		for _, fill := range fills {
			trade, err := c.convertFillToTrade(fill)
			if err != nil {
				log.Printf("Warning: Failed to convert fill: %v", err)
				continue
			}
			allTrades = append(allTrades, trade)
		}

		// If we got less than max batch size, we've reached the end
		if len(fills) < config.MaxTradesPerBatch {
			log.Printf("Fetched %d trades in %d batches", len(allTrades), batchCount)
			break
		}

		// Update start time to the timestamp of the last fill + 1ms for next batch
		lastFillTime := fills[len(fills)-1].Time
		currentStartTime = lastFillTime + 1
	}

	return allTrades, nil
}

// fetchBatch fetches a single batch of trades from the API
func (c *HyperliquidClient) fetchBatch(address string, startTime, endTime int64) ([]FillResponse, error) {
	requestBody := UserFillsRequest{
		Type:            "userFillsByTime",
		User:            address,
		StartTime:       &startTime,
		EndTime:         &endTime,
		AggregateByTime: true,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(config.HyperliquidAPIURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trades: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var fills []FillResponse
	if err := json.Unmarshal(body, &fills); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return fills, nil
}

// convertFillToTrade converts a FillResponse to a Trade model
func (c *HyperliquidClient) convertFillToTrade(fill FillResponse) (models.Trade, error) {
	price, err := strconv.ParseFloat(fill.Price, 64)
	if err != nil {
		return models.Trade{}, fmt.Errorf("failed to parse price '%s': %w", fill.Price, err)
	}

	size, err := strconv.ParseFloat(fill.Size, 64)
	if err != nil {
		return models.Trade{}, fmt.Errorf("failed to parse size '%s': %w", fill.Size, err)
	}

	return models.Trade{
		Time:  time.UnixMilli(fill.Time),
		Coin:  fill.Coin,
		Side:  fill.Side,
		Price: price,
		Size:  size,
		Value: price * size,
	}, nil
}
