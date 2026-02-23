package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// MarketData 代表由交易所抓取回來的市場數據
type MarketData struct {
	Symbol       string
	Price        decimal.Decimal
	FundingRate  decimal.Decimal
	OpenInterest decimal.Decimal
	Timestamp    time.Time
}

// AlertEvent 代表策略觸發的告警事件
type AlertEvent struct {
	Symbol       string
	Price        decimal.Decimal
	FundingRate  decimal.Decimal
	CurrentOI    decimal.Decimal
	RecentAvgOI  decimal.Decimal
	OlderAvgOI   decimal.Decimal
	OISurgeRatio decimal.Decimal
	Timestamp    time.Time
}
