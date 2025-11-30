package models

import "time"

type Trade struct {
	Time      time.Time `json:"time"`
	Coin      string    `json:"coin"`
	Side      string    `json:"side"` // "B" for buy, "A" for sell
	Price     float64   `json:"px"`
	Size      float64   `json:"sz"`
	Value     float64   `json:"value"`
}

type DailyPnL struct {
	Date          string  `json:"date"`
	TradeCount    int     `json:"tradeCount"`
	DailyPnL      float64 `json:"dailyPnL"`
	CumulativePnL float64 `json:"cumulativePnL"`
}

type PnLSummary struct {
	DailyRecords []DailyPnL `json:"dailyRecords"`
	TotalPnL     float64    `json:"totalPnL"`
}