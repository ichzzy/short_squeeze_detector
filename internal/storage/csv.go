package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ichzzy/short-squeeze-detector/internal/model"
)

// CSVStorage 負責寫入資料到 CSV
type CSVStorage struct {
	dir string
}

// NewCSVStorage 建立 CSVStorage 實例
func NewCSVStorage(dir string) *CSVStorage {
	os.MkdirAll(dir, 0755)
	return &CSVStorage{dir: dir}
}

// Append 附加一筆市價到對應幣對的 CSV，如果檔案不存在則加入 Header
func (s *CSVStorage) Append(data *model.MarketData) error {
	filePath := filepath.Join(s.dir, fmt.Sprintf("%s.csv", data.Symbol))

	// 判斷檔案是否存在
	addHeader := false
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		addHeader = true
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	if addHeader {
		if err := writer.Write([]string{"Timestamp", "Price", "FundingRate", "OpenInterest"}); err != nil {
			return err
		}
	}

	row := []string{
		data.Timestamp.Format("2006-01-02 15:04:05"),
		data.Price.String(),
		data.FundingRate.String(),
		data.OpenInterest.String(),
	}

	return writer.Write(row)
}
