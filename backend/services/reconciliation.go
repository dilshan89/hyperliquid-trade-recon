package services

import (
	"fmt"
	"hyperliquid-recon/models"
	"log"
	"sort"
	"sync"
	"time"
)

// AccountCache stores cached data for a specific account
type AccountCache struct {
	trades        []models.Trade
	lastFetchTime time.Time
	cachedDays    int // Maximum days of data we have in cache
}

// ReconciliationService handles trade reconciliation and P&L calculations
type ReconciliationService struct {
	accountCache map[string]*AccountCache // key: address
	dailyPnL     map[string]*models.DailyPnL
	mu           sync.RWMutex
	hlClient     *HyperliquidClient
}

// NewReconciliationService creates a new reconciliation service
func NewReconciliationService() *ReconciliationService {
	return &ReconciliationService{
		accountCache: make(map[string]*AccountCache),
		dailyPnL:     make(map[string]*models.DailyPnL),
		hlClient:     NewHyperliquidClient(),
	}
}

// FetchAndReconcile fetches trades for an address and calculates P&L
// Uses intelligent caching: incremental fetch for same range, cache reuse for smaller range
func (rs *ReconciliationService) FetchAndReconcile(address string, days int) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	cache, exists := rs.accountCache[address]
	now := time.Now()

	if exists && !cache.lastFetchTime.IsZero() {
		timeSinceLastFetch := now.Sub(cache.lastFetchTime)

		// Case 1: Requesting SMALLER time range than cached (e.g., 7D when we have 30D)
		if days <= cache.cachedDays && timeSinceLastFetch < time.Hour {
			log.Printf("Cache reuse for %s: requested %d days, have %d days cached", address, days, cache.cachedDays)

			// Fetch only new trades since last fetch
			newTrades, err := rs.hlClient.FetchTradesInRange(address, cache.lastFetchTime, now)
			if err != nil {
				return err
			}

			if len(newTrades) > 0 {
				log.Printf("Found %d new trades, merging with %d cached trades", len(newTrades), len(cache.trades))
				cache.trades = rs.mergeTrades(cache.trades, newTrades)
			} else {
				log.Printf("No new trades found, using cached trades")
			}

			// Update last fetch time (keep original cachedDays)
			cache.lastFetchTime = now

			// Filter trades to requested time range
			cutoffTime := now.Add(-time.Duration(days) * 24 * time.Hour)
			filteredTrades := rs.filterTradesByTime(cache.trades, cutoffTime)

			log.Printf("Filtered %d trades to %d trades for %d days", len(cache.trades), len(filteredTrades), days)

			// Calculate P&L from filtered trades
			rs.calculateDailyPnLFromTrades(filteredTrades)

			log.Printf("Cache reuse complete: %d trades, %d days", len(filteredTrades), len(rs.dailyPnL))
			return nil
		}

		// Case 2: Requesting SAME time range as cached
		if days == cache.cachedDays && timeSinceLastFetch < time.Hour {
			log.Printf("Incremental fetch for %s: fetching new trades since %s", address, cache.lastFetchTime.Format(time.RFC3339))

			// Fetch only new trades since last fetch
			newTrades, err := rs.hlClient.FetchTradesInRange(address, cache.lastFetchTime, now)
			if err != nil {
				return err
			}

			if len(newTrades) > 0 {
				log.Printf("Found %d new trades, merging with %d cached trades", len(newTrades), len(cache.trades))
				cache.trades = rs.mergeTrades(cache.trades, newTrades)
			} else {
				log.Printf("No new trades found, using cached %d trades", len(cache.trades))
			}

			// Update last fetch time
			cache.lastFetchTime = now

			// Calculate P&L from cached trades
			rs.calculateDailyPnLFromTrades(cache.trades)

			log.Printf("Incremental reconciliation complete: %d total trades, %d days", len(cache.trades), len(rs.dailyPnL))
			return nil
		}
	}

	// Case 3: Full fetch needed (no cache, larger range requested, or cache too old)
	log.Printf("Full fetch for %s: fetching all trades for last %d days", address, days)

	trades, err := rs.hlClient.FetchTrades(address, days)
	if err != nil {
		return err
	}

	// Create or update cache
	rs.accountCache[address] = &AccountCache{
		trades:        trades,
		lastFetchTime: now,
		cachedDays:    days,
	}

	rs.calculateDailyPnLFromTrades(trades)

	log.Printf("Full reconciliation complete: %d trades, %d days", len(trades), len(rs.dailyPnL))

	return nil
}

// filterTradesByTime filters trades to only include those after the cutoff time
func (rs *ReconciliationService) filterTradesByTime(trades []models.Trade, cutoffTime time.Time) []models.Trade {
	filtered := make([]models.Trade, 0, len(trades))
	for _, trade := range trades {
		if trade.Time.After(cutoffTime) || trade.Time.Equal(cutoffTime) {
			filtered = append(filtered, trade)
		}
	}
	return filtered
}

// mergeTrades combines existing and new trades, removing duplicates
func (rs *ReconciliationService) mergeTrades(existing, new []models.Trade) []models.Trade {
	// Use a map to track unique trades by timestamp+coin+side to avoid duplicates
	tradeMap := make(map[string]models.Trade)

	// Add existing trades to map
	for _, trade := range existing {
		key := fmt.Sprintf("%d_%s_%s", trade.Time.UnixMilli(), trade.Coin, trade.Side)
		tradeMap[key] = trade
	}

	// Add new trades (will overwrite if duplicate)
	for _, trade := range new {
		key := fmt.Sprintf("%d_%s_%s", trade.Time.UnixMilli(), trade.Coin, trade.Side)
		tradeMap[key] = trade
	}

	// Convert map back to slice
	merged := make([]models.Trade, 0, len(tradeMap))
	for _, trade := range tradeMap {
		merged = append(merged, trade)
	}

	// Sort by time ascending
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Time.Before(merged[j].Time)
	})

	return merged
}

// calculateDailyPnLFromTrades groups trades by date and calculates daily P&L
func (rs *ReconciliationService) calculateDailyPnLFromTrades(trades []models.Trade) {
	// Group trades by date
	tradesByDate := make(map[string][]models.Trade)

	for _, trade := range trades {
		dateKey := trade.Time.Format("2006-01-02") // Go's reference date format
		tradesByDate[dateKey] = append(tradesByDate[dateKey], trade)
	}

	// Calculate P&L for each day
	rs.dailyPnL = make(map[string]*models.DailyPnL)

	for date, dayTrades := range tradesByDate {
		pnl := rs.calculatePnLForDay(dayTrades)
		rs.dailyPnL[date] = &models.DailyPnL{
			Date:       date,
			TradeCount: len(dayTrades),
			DailyPnL:   pnl,
		}
	}
}

// calculatePnLForDay calculates P&L for a single day's trades
// P&L = Total Sell Value - Total Buy Value, grouped by coin
func (rs *ReconciliationService) calculatePnLForDay(trades []models.Trade) float64 {
	coinPositions := make(map[string]*Position)

	for _, trade := range trades {
		if _, exists := coinPositions[trade.Coin]; !exists {
			coinPositions[trade.Coin] = &Position{
				BuyValue:  0,
				SellValue: 0,
			}
		}

		if trade.Side == "B" {
			coinPositions[trade.Coin].BuyValue += trade.Value
		} else if trade.Side == "A" {
			coinPositions[trade.Coin].SellValue += trade.Value
		}
	}

	totalPnL := 0.0
	for _, pos := range coinPositions {
		totalPnL += pos.SellValue - pos.BuyValue
	}

	return totalPnL
}

// Position tracks buy and sell values for a coin
type Position struct {
	BuyValue  float64
	SellValue float64
}

// GetPnLSummary returns a summary of all P&L calculations
func (rs *ReconciliationService) GetPnLSummary() models.PnLSummary {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	// Convert map to sorted slice
	records := make([]models.DailyPnL, 0, len(rs.dailyPnL))
	for _, record := range rs.dailyPnL {
		records = append(records, *record)
	}

	// Sort by date descending
	sort.Slice(records, func(i, j int) bool {
		return records[i].Date > records[j].Date
	})

	// Calculate cumulative P&L
	cumulative := 0.0
	for i := len(records) - 1; i >= 0; i-- {
		cumulative += records[i].DailyPnL
		records[i].CumulativePnL = cumulative
	}

	totalPnL := 0.0
	for _, record := range records {
		totalPnL += record.DailyPnL
	}

	return models.PnLSummary{
		DailyRecords: records,
		TotalPnL:     totalPnL,
	}
}
