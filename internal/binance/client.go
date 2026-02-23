package binance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ichzzy/short-squeeze-detector/internal/model"
	"github.com/shopspring/decimal"
)

const (
	baseURL = "https://fapi.binance.com"
)

// Client Binance Client
type Client struct {
	httpClient *http.Client
	apiKey     string
	apiSecret  string
}

// NewClient 建立 Binance API Client
func NewClient(apiKey, apiSecret string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiKey:     apiKey,
		apiSecret:  apiSecret,
	}
}

// premiumIndexResponse 接收資金費率等資訊
type premiumIndexResponse struct {
	Symbol          string `json:"symbol"`
	MarkPrice       string `json:"markPrice"`
	LastFundingRate string `json:"lastFundingRate"`
	Time            int64  `json:"time"`
}

// openInterestResponse 接收持倉量資訊
type openInterestResponse struct {
	Symbol       string `json:"symbol"`
	OpenInterest string `json:"openInterest"`
	Time         int64  `json:"time"`
}

// FetchMarketData 抓取指定幣對的市場行情與資金費率、OI
func (c *Client) FetchMarketData(symbol string) (*model.MarketData, error) {
	// 1. 取得 Premium Index (包含 MarkPrice & FundingRate)
	premiumURL := fmt.Sprintf("%s/fapi/v1/premiumIndex?symbol=%s", baseURL, symbol)
	respP, err := c.httpClient.Get(premiumURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get premium index: %v", err)
	}
	defer respP.Body.Close()
	if respP.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("premium index API returned status %d", respP.StatusCode)
	}
	var pData premiumIndexResponse
	if err := json.NewDecoder(respP.Body).Decode(&pData); err != nil {
		return nil, err
	}

	// 2. 取得 Open Interest
	oiURL := fmt.Sprintf("%s/fapi/v1/openInterest?symbol=%s", baseURL, symbol)
	respO, err := c.httpClient.Get(oiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get open interest: %v", err)
	}
	defer respO.Body.Close()
	if respO.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("open interest API returned status %d", respO.StatusCode)
	}
	var oData openInterestResponse
	if err := json.NewDecoder(respO.Body).Decode(&oData); err != nil {
		return nil, err
	}

	price, _ := decimal.NewFromString(pData.MarkPrice)
	fundingRate, _ := decimal.NewFromString(pData.LastFundingRate)
	oi, _ := decimal.NewFromString(oData.OpenInterest)

	return &model.MarketData{
		Symbol:       symbol,
		Price:        price,
		FundingRate:  fundingRate,
		OpenInterest: oi,
		Timestamp:    time.UnixMilli(pData.Time),
	}, nil
}
