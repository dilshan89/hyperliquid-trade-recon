package services

import (
	"hyperliquid-recon/models"
	"testing"
	"time"
)

// Helper function to create a test trade
func createTestTrade(timeStr string, coin string, side string, price float64, size float64) models.Trade {
	t, _ := time.Parse(time.RFC3339, timeStr)
	return models.Trade{
		Time:  t,
		Coin:  coin,
		Side:  side,
		Price: price,
		Size:  size,
		Value: price * size,
	}
}

// Test filterTradesByTime
func TestFilterTradesByTime(t *testing.T) {
	rs := NewReconciliationService()

	t.Run("should include trades after cutoff", func(t *testing.T) {
		trades := []models.Trade{
			createTestTrade("2025-01-01T10:00:00Z", "BTC", "B", 50000, 1),
			createTestTrade("2025-01-02T10:00:00Z", "BTC", "A", 51000, 1),
			createTestTrade("2025-01-03T10:00:00Z", "ETH", "B", 3000, 2),
		}

		cutoff, _ := time.Parse(time.RFC3339, "2025-01-02T00:00:00Z")
		filtered := rs.filterTradesByTime(trades, cutoff)

		if len(filtered) != 2 {
			t.Errorf("Expected 2 trades, got %d", len(filtered))
		}
	})

	t.Run("should include trades equal to cutoff", func(t *testing.T) {
		trades := []models.Trade{
			createTestTrade("2025-01-01T10:00:00Z", "BTC", "B", 50000, 1),
			createTestTrade("2025-01-02T00:00:00Z", "BTC", "A", 51000, 1),
		}

		cutoff, _ := time.Parse(time.RFC3339, "2025-01-02T00:00:00Z")
		filtered := rs.filterTradesByTime(trades, cutoff)

		if len(filtered) != 1 {
			t.Errorf("Expected 1 trade, got %d", len(filtered))
		}
		if filtered[0].Coin != "BTC" || filtered[0].Side != "A" {
			t.Errorf("Wrong trade included")
		}
	})

	t.Run("should handle empty slice", func(t *testing.T) {
		trades := []models.Trade{}
		cutoff, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00Z")
		filtered := rs.filterTradesByTime(trades, cutoff)

		if len(filtered) != 0 {
			t.Errorf("Expected 0 trades, got %d", len(filtered))
		}
	})

	t.Run("should exclude all trades before cutoff", func(t *testing.T) {
		trades := []models.Trade{
			createTestTrade("2025-01-01T10:00:00Z", "BTC", "B", 50000, 1),
			createTestTrade("2025-01-01T11:00:00Z", "BTC", "A", 51000, 1),
		}

		cutoff, _ := time.Parse(time.RFC3339, "2025-01-02T00:00:00Z")
		filtered := rs.filterTradesByTime(trades, cutoff)

		if len(filtered) != 0 {
			t.Errorf("Expected 0 trades, got %d", len(filtered))
		}
	})
}

// Test mergeTrades
func TestMergeTrades(t *testing.T) {
	rs := NewReconciliationService()

	t.Run("should merge non-overlapping trades", func(t *testing.T) {
		existing := []models.Trade{
			createTestTrade("2025-01-01T10:00:00Z", "BTC", "B", 50000, 1),
		}
		new := []models.Trade{
			createTestTrade("2025-01-02T10:00:00Z", "BTC", "A", 51000, 1),
		}

		merged := rs.mergeTrades(existing, new)

		if len(merged) != 2 {
			t.Errorf("Expected 2 trades, got %d", len(merged))
		}

		// Check sorting - should be ascending by time
		if !merged[0].Time.Before(merged[1].Time) {
			t.Errorf("Trades not sorted correctly")
		}
	})

	t.Run("should deduplicate identical trades", func(t *testing.T) {
		existing := []models.Trade{
			createTestTrade("2025-01-01T10:00:00Z", "BTC", "B", 50000, 1),
		}
		new := []models.Trade{
			createTestTrade("2025-01-01T10:00:00Z", "BTC", "B", 50000, 1),
		}

		merged := rs.mergeTrades(existing, new)

		if len(merged) != 1 {
			t.Errorf("Expected 1 trade (deduplicated), got %d", len(merged))
		}
	})

	t.Run("should handle empty existing trades", func(t *testing.T) {
		existing := []models.Trade{}
		new := []models.Trade{
			createTestTrade("2025-01-01T10:00:00Z", "BTC", "B", 50000, 1),
		}

		merged := rs.mergeTrades(existing, new)

		if len(merged) != 1 {
			t.Errorf("Expected 1 trade, got %d", len(merged))
		}
	})

	t.Run("should handle empty new trades", func(t *testing.T) {
		existing := []models.Trade{
			createTestTrade("2025-01-01T10:00:00Z", "BTC", "B", 50000, 1),
		}
		new := []models.Trade{}

		merged := rs.mergeTrades(existing, new)

		if len(merged) != 1 {
			t.Errorf("Expected 1 trade, got %d", len(merged))
		}
	})

	t.Run("should sort merged trades by time ascending", func(t *testing.T) {
		existing := []models.Trade{
			createTestTrade("2025-01-03T10:00:00Z", "BTC", "B", 50000, 1),
		}
		new := []models.Trade{
			createTestTrade("2025-01-01T10:00:00Z", "ETH", "A", 3000, 1),
			createTestTrade("2025-01-02T10:00:00Z", "BTC", "A", 51000, 1),
		}

		merged := rs.mergeTrades(existing, new)

		if len(merged) != 3 {
			t.Errorf("Expected 3 trades, got %d", len(merged))
		}

		// Verify ascending order
		for i := 0; i < len(merged)-1; i++ {
			if !merged[i].Time.Before(merged[i+1].Time) {
				t.Errorf("Trades not sorted in ascending order at index %d", i)
			}
		}
	})
}

// Test calculatePnLForDay
func TestCalculatePnLForDay(t *testing.T) {
	rs := NewReconciliationService()

	t.Run("should calculate positive P&L", func(t *testing.T) {
		trades := []models.Trade{
			createTestTrade("2025-01-01T10:00:00Z", "BTC", "B", 50000, 1),  // Buy 1 BTC at 50000 = -50000
			createTestTrade("2025-01-01T11:00:00Z", "BTC", "A", 51000, 1),  // Sell 1 BTC at 51000 = +51000
		}

		pnl := rs.calculatePnLForDay(trades)

		expected := 1000.0 // 51000 - 50000
		if pnl != expected {
			t.Errorf("Expected P&L %f, got %f", expected, pnl)
		}
	})

	t.Run("should calculate negative P&L", func(t *testing.T) {
		trades := []models.Trade{
			createTestTrade("2025-01-01T10:00:00Z", "BTC", "B", 51000, 1),  // Buy 1 BTC at 51000 = -51000
			createTestTrade("2025-01-01T11:00:00Z", "BTC", "A", 50000, 1),  // Sell 1 BTC at 50000 = +50000
		}

		pnl := rs.calculatePnLForDay(trades)

		expected := -1000.0 // 50000 - 51000
		if pnl != expected {
			t.Errorf("Expected P&L %f, got %f", expected, pnl)
		}
	})

	t.Run("should handle multiple coins", func(t *testing.T) {
		trades := []models.Trade{
			createTestTrade("2025-01-01T10:00:00Z", "BTC", "B", 50000, 1),  // Buy BTC: -50000
			createTestTrade("2025-01-01T11:00:00Z", "BTC", "A", 51000, 1),  // Sell BTC: +51000, P&L: +1000
			createTestTrade("2025-01-01T12:00:00Z", "ETH", "B", 3000, 2),   // Buy ETH: -6000
			createTestTrade("2025-01-01T13:00:00Z", "ETH", "A", 3100, 2),   // Sell ETH: +6200, P&L: +200
		}

		pnl := rs.calculatePnLForDay(trades)

		expected := 1200.0 // BTC: +1000, ETH: +200
		if pnl != expected {
			t.Errorf("Expected P&L %f, got %f", expected, pnl)
		}
	})

	t.Run("should handle empty trades", func(t *testing.T) {
		trades := []models.Trade{}
		pnl := rs.calculatePnLForDay(trades)

		if pnl != 0 {
			t.Errorf("Expected P&L 0, got %f", pnl)
		}
	})

	t.Run("should handle only buy trades", func(t *testing.T) {
		trades := []models.Trade{
			createTestTrade("2025-01-01T10:00:00Z", "BTC", "B", 50000, 1),
		}

		pnl := rs.calculatePnLForDay(trades)

		expected := -50000.0 // Only buys, negative P&L
		if pnl != expected {
			t.Errorf("Expected P&L %f, got %f", expected, pnl)
		}
	})

	t.Run("should handle only sell trades", func(t *testing.T) {
		trades := []models.Trade{
			createTestTrade("2025-01-01T10:00:00Z", "BTC", "A", 50000, 1),
		}

		pnl := rs.calculatePnLForDay(trades)

		expected := 50000.0 // Only sells, positive P&L
		if pnl != expected {
			t.Errorf("Expected P&L %f, got %f", expected, pnl)
		}
	})
}

// Test calculateDailyPnLFromTrades
func TestCalculateDailyPnLFromTrades(t *testing.T) {
	rs := NewReconciliationService()

	t.Run("should group trades by date", func(t *testing.T) {
		trades := []models.Trade{
			createTestTrade("2025-01-01T10:00:00Z", "BTC", "B", 50000, 1),
			createTestTrade("2025-01-01T23:00:00Z", "BTC", "A", 51000, 1),
			createTestTrade("2025-01-02T10:00:00Z", "ETH", "B", 3000, 1),
			createTestTrade("2025-01-02T12:00:00Z", "ETH", "A", 3100, 1),
		}

		rs.calculateDailyPnLFromTrades(trades)

		if len(rs.dailyPnL) != 2 {
			t.Errorf("Expected 2 days, got %d", len(rs.dailyPnL))
		}

		// Check 2025-01-01
		day1, exists := rs.dailyPnL["2025-01-01"]
		if !exists {
			t.Error("Expected data for 2025-01-01")
		}
		if day1.TradeCount != 2 {
			t.Errorf("Expected 2 trades on 2025-01-01, got %d", day1.TradeCount)
		}
		if day1.DailyPnL != 1000.0 {
			t.Errorf("Expected P&L 1000 on 2025-01-01, got %f", day1.DailyPnL)
		}

		// Check 2025-01-02
		day2, exists := rs.dailyPnL["2025-01-02"]
		if !exists {
			t.Error("Expected data for 2025-01-02")
		}
		if day2.TradeCount != 2 {
			t.Errorf("Expected 2 trades on 2025-01-02, got %d", day2.TradeCount)
		}
		if day2.DailyPnL != 100.0 {
			t.Errorf("Expected P&L 100 on 2025-01-02, got %f", day2.DailyPnL)
		}
	})

	t.Run("should handle trades across midnight", func(t *testing.T) {
		trades := []models.Trade{
			createTestTrade("2025-01-01T23:59:00Z", "BTC", "B", 50000, 1),
			createTestTrade("2025-01-02T00:01:00Z", "BTC", "A", 51000, 1),
		}

		rs.calculateDailyPnLFromTrades(trades)

		if len(rs.dailyPnL) != 2 {
			t.Errorf("Expected 2 days, got %d", len(rs.dailyPnL))
		}
	})

	t.Run("should handle empty trades", func(t *testing.T) {
		trades := []models.Trade{}
		rs.calculateDailyPnLFromTrades(trades)

		if len(rs.dailyPnL) != 0 {
			t.Errorf("Expected 0 days, got %d", len(rs.dailyPnL))
		}
	})
}

// Test GetPnLSummary
func TestGetPnLSummary(t *testing.T) {
	rs := NewReconciliationService()

	t.Run("should return sorted records descending by date", func(t *testing.T) {
		trades := []models.Trade{
			createTestTrade("2025-01-01T10:00:00Z", "BTC", "B", 50000, 1),
			createTestTrade("2025-01-01T11:00:00Z", "BTC", "A", 51000, 1),
			createTestTrade("2025-01-02T10:00:00Z", "ETH", "B", 3000, 1),
			createTestTrade("2025-01-02T11:00:00Z", "ETH", "A", 3100, 1),
			createTestTrade("2025-01-03T10:00:00Z", "BTC", "B", 52000, 1),
			createTestTrade("2025-01-03T11:00:00Z", "BTC", "A", 52500, 1),
		}

		rs.calculateDailyPnLFromTrades(trades)
		summary := rs.GetPnLSummary()

		if len(summary.DailyRecords) != 3 {
			t.Errorf("Expected 3 records, got %d", len(summary.DailyRecords))
		}

		// Check descending order
		if summary.DailyRecords[0].Date != "2025-01-03" {
			t.Errorf("Expected first record to be 2025-01-03, got %s", summary.DailyRecords[0].Date)
		}
		if summary.DailyRecords[1].Date != "2025-01-02" {
			t.Errorf("Expected second record to be 2025-01-02, got %s", summary.DailyRecords[1].Date)
		}
		if summary.DailyRecords[2].Date != "2025-01-01" {
			t.Errorf("Expected third record to be 2025-01-01, got %s", summary.DailyRecords[2].Date)
		}
	})

	t.Run("should calculate cumulative P&L correctly", func(t *testing.T) {
		trades := []models.Trade{
			// Day 1: +1000
			createTestTrade("2025-01-01T10:00:00Z", "BTC", "B", 50000, 1),
			createTestTrade("2025-01-01T11:00:00Z", "BTC", "A", 51000, 1),
			// Day 2: +100
			createTestTrade("2025-01-02T10:00:00Z", "ETH", "B", 3000, 1),
			createTestTrade("2025-01-02T11:00:00Z", "ETH", "A", 3100, 1),
			// Day 3: +500
			createTestTrade("2025-01-03T10:00:00Z", "BTC", "B", 52000, 1),
			createTestTrade("2025-01-03T11:00:00Z", "BTC", "A", 52500, 1),
		}

		rs.calculateDailyPnLFromTrades(trades)
		summary := rs.GetPnLSummary()

		// Cumulative should be calculated from earliest to latest
		// But displayed in descending order
		// Day 1: cumulative = 1000
		// Day 2: cumulative = 1100
		// Day 3: cumulative = 1600

		// Find each day's record
		var day1, day2, day3 *models.DailyPnL
		for i := range summary.DailyRecords {
			switch summary.DailyRecords[i].Date {
			case "2025-01-01":
				day1 = &summary.DailyRecords[i]
			case "2025-01-02":
				day2 = &summary.DailyRecords[i]
			case "2025-01-03":
				day3 = &summary.DailyRecords[i]
			}
		}

		if day1 == nil || day2 == nil || day3 == nil {
			t.Fatal("Missing expected date records")
		}

		if day1.CumulativePnL != 1000.0 {
			t.Errorf("Expected cumulative P&L 1000 for day 1, got %f", day1.CumulativePnL)
		}
		if day2.CumulativePnL != 1100.0 {
			t.Errorf("Expected cumulative P&L 1100 for day 2, got %f", day2.CumulativePnL)
		}
		if day3.CumulativePnL != 1600.0 {
			t.Errorf("Expected cumulative P&L 1600 for day 3, got %f", day3.CumulativePnL)
		}
	})

	t.Run("should calculate total P&L correctly", func(t *testing.T) {
		trades := []models.Trade{
			createTestTrade("2025-01-01T10:00:00Z", "BTC", "B", 50000, 1),
			createTestTrade("2025-01-01T11:00:00Z", "BTC", "A", 51000, 1),  // +1000
			createTestTrade("2025-01-02T10:00:00Z", "ETH", "B", 3100, 1),
			createTestTrade("2025-01-02T11:00:00Z", "ETH", "A", 3000, 1),   // -100
		}

		rs.calculateDailyPnLFromTrades(trades)
		summary := rs.GetPnLSummary()

		expected := 900.0 // 1000 + (-100)
		if summary.TotalPnL != expected {
			t.Errorf("Expected total P&L %f, got %f", expected, summary.TotalPnL)
		}
	})

	t.Run("should handle negative total P&L", func(t *testing.T) {
		trades := []models.Trade{
			createTestTrade("2025-01-01T10:00:00Z", "BTC", "B", 51000, 1),
			createTestTrade("2025-01-01T11:00:00Z", "BTC", "A", 50000, 1),  // -1000
		}

		rs.calculateDailyPnLFromTrades(trades)
		summary := rs.GetPnLSummary()

		expected := -1000.0
		if summary.TotalPnL != expected {
			t.Errorf("Expected total P&L %f, got %f", expected, summary.TotalPnL)
		}
	})

	t.Run("should handle empty data", func(t *testing.T) {
		trades := []models.Trade{}
		rs.calculateDailyPnLFromTrades(trades)
		summary := rs.GetPnLSummary()

		if len(summary.DailyRecords) != 0 {
			t.Errorf("Expected 0 records, got %d", len(summary.DailyRecords))
		}
		if summary.TotalPnL != 0 {
			t.Errorf("Expected total P&L 0, got %f", summary.TotalPnL)
		}
	})
}