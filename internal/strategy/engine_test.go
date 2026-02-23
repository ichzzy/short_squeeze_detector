package strategy

import (
	"testing"
	"time"

	"github.com/ichzzy/short-squeeze-detector/internal/model"
	"github.com/shopspring/decimal"
)

func TestEngine_Process(t *testing.T) {
	engine := NewEngine(0.001, 2.0)
	symbol := "BTCUSDT"

	// 準備 7 個正常數據以累積
	for i := 0; i < 7; i++ {
		res := engine.Process(&model.MarketData{
			Symbol:       symbol,
			Price:        decimal.NewFromInt(10000),
			FundingRate:  decimal.NewFromFloat(0.0001),
			OpenInterest: decimal.NewFromInt(1000),
			Timestamp:    time.Now(),
		})
		if res != nil {
			t.Errorf("Expect nil before 10 records, got %v", res)
		}
	}

	// 新增 3 個極端激增的數據
	// 第8筆
	engine.Process(&model.MarketData{
		Symbol:       symbol,
		Price:        decimal.NewFromInt(10100),
		FundingRate:  decimal.NewFromFloat(0.0001), // 還沒超過費率閾值
		OpenInterest: decimal.NewFromInt(2500),
		Timestamp:    time.Now(),
	})

	// 第9筆
	engine.Process(&model.MarketData{
		Symbol:       symbol,
		Price:        decimal.NewFromInt(10200),
		FundingRate:  decimal.NewFromFloat(-0.0015), // 費率超過閾值了
		OpenInterest: decimal.NewFromInt(2600),
		Timestamp:    time.Now(),
	})

	// 第10筆，湊滿 10 筆，且此時最後三筆 OI 分別為 2500, 2600, 2700 (avg=2600)
	// 全部 10 筆中：7筆 * 1000 = 7000，加上 2500, 2600, 2700 = 7800 -> 總和 14800, avg=1480
	// 激增倍率 = 2600 / 1480 = 1.75... < 2.0 因此不應觸發
	ev1 := engine.Process(&model.MarketData{
		Symbol:       symbol,
		Price:        decimal.NewFromInt(10300),
		FundingRate:  decimal.NewFromFloat(-0.0015),
		OpenInterest: decimal.NewFromInt(2700),
		Timestamp:    time.Now(),
	})

	if ev1 != nil {
		t.Errorf("Should not alert, ratio is below 2.0. Expected nil, got %v", ev1)
	}

	// 第11筆 再加入更大的一筆，把倍率推過 2.0
	// 剔除最早的一筆 (1000)，新的最後三筆: 2600, 2700, 5000 (avg=3433.33)
	// 全部 10 筆: 6*1000 + 2500 + 2600 + 2700 + 5000 = 18800 (avg=1880)
	// 激增 = 3433.33 / 1880 = 1.82... 還是不夠
	// 我們把這筆設超級大
	ev2 := engine.Process(&model.MarketData{
		Symbol:       symbol,
		Price:        decimal.NewFromInt(10400),
		FundingRate:  decimal.NewFromFloat(-0.002),
		OpenInterest: decimal.NewFromInt(15000),
		Timestamp:    time.Now(),
	})

	if ev2 == nil {
		t.Errorf("Expected alert for huge surge, got nil")
	} else if ev2.OISurgeRatio.LessThanOrEqual(decimal.NewFromFloat(2.0)) {
		t.Errorf("Expected OISurgeRatio > 2.0, got %s", ev2.OISurgeRatio.String())
	}
}
