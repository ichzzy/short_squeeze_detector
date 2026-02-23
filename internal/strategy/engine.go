package strategy

import (
	"log"
	"sync"
	"time"

	"github.com/ichzzy/short-squeeze-detector/internal/model"
	"github.com/shopspring/decimal"
)

const (
	averageWindow = 10
	recentWindow  = 3
)

// 策略引擎
type Engine struct {
	fundingRateThreshold decimal.Decimal
	oiSurgeRatio         decimal.Decimal

	// symbol -> []MarketData
	history map[string][]model.MarketData
	mu      sync.Mutex
}

func NewEngine(fundingThreshold, oiSurgeRatio float64) *Engine {
	return &Engine{
		fundingRateThreshold: decimal.NewFromFloat(fundingThreshold).Abs(),
		oiSurgeRatio:         decimal.NewFromFloat(oiSurgeRatio),
		history:              make(map[string][]model.MarketData),
	}
}

func (e *Engine) Process(data *model.MarketData) *model.AlertEvent {
	e.mu.Lock()
	defer e.mu.Unlock()

	records := e.history[data.Symbol]
	records = append(records, *data)

	// 保留最近 M 次
	if len(records) > averageWindow {
		records = records[len(records)-averageWindow:]
	}
	e.history[data.Symbol] = records

	// 若紀錄不足 M 筆，先不進行計算
	if len(records) < averageWindow {
		return nil
	}

	// 條件一：資金費率絕對值大於閾值
	if data.FundingRate.Abs().LessThan(e.fundingRateThreshold) {
		return nil
	}

	// 條件二：最近 N 次的OI均值 / 最近 M 次的 OI均值 > ratio
	sumRecent := decimal.NewFromInt(0)
	for i := len(records) - recentWindow; i < len(records); i++ {
		sumRecent = sumRecent.Add(records[i].OpenInterest)
	}
	avgRecent := sumRecent.Div(decimal.NewFromInt(recentWindow))

	sumAll := decimal.NewFromInt(0)
	for i := 0; i < len(records); i++ {
		sumAll = sumAll.Add(records[i].OpenInterest)
	}
	avgAll := sumAll.Div(decimal.NewFromInt(averageWindow))

	if avgAll.IsZero() {
		return nil // 避免除以零
	}

	surge := avgRecent.Div(avgAll)
	if surge.GreaterThan(e.oiSurgeRatio) {
		log.Printf("[%s] Short squeeze alert triggered: FundingRate=%s, OISurge=%s", data.Symbol, data.FundingRate.String(), surge.StringFixed(2))
		return &model.AlertEvent{
			Symbol:       data.Symbol,
			Price:        data.Price,
			FundingRate:  data.FundingRate,
			CurrentOI:    data.OpenInterest,
			RecentAvgOI:  avgRecent,
			OlderAvgOI:   avgAll,
			OISurgeRatio: surge,
			Timestamp:    time.Now(),
		}
	}

	return nil
}
