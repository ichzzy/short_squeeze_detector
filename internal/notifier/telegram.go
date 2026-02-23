package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ichzzy/short-squeeze-detector/internal/model"
	"github.com/shopspring/decimal"
)

// TelegramNotifier 傳送通知到 Telegram
type TelegramNotifier struct {
	token  string
	chatID string
	client *http.Client
}

// NewTelegramNotifier 建立 Notifier
func NewTelegramNotifier(token, chatID string) *TelegramNotifier {
	return &TelegramNotifier{
		token:  token,
		chatID: chatID,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// SendAlert 將策略觸發的告警轉為文字傳送
func (t *TelegramNotifier) SendAlert(event *model.AlertEvent) error {
	if t.token == "" || t.chatID == "" {
		// 未配置 Telegram，略過
		return nil
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)

	// Markdown V2 格式
	msgText := fmt.Sprintf(
		"🚨 *山寨幣軋空警報* 🚨\n\n"+
			"*幣對:* %s\n"+
			"*當前價格:* %.4f\n"+
			"*資金費率:* %.4f%%\n"+
			"*OI 短期激增比例:* %.2f 倍\n"+
			"_(近期 OI 平均: %.2f, 過去 OI 平均: %.2f)_\n"+
			"*時間:* %s\n\n"+
			"⚠️  _請注意流動性風險與技術分析止盈止損_",
		event.Symbol,
		event.Price.InexactFloat64(),
		event.FundingRate.Mul(decimal.NewFromInt(100)).InexactFloat64(), // 轉為 %
		event.OISurgeRatio.InexactFloat64(),
		event.RecentAvgOI.InexactFloat64(),
		event.OlderAvgOI.InexactFloat64(),
		event.Timestamp.Format("2006-01-02 15:04:05"),
	)

	payload := map[string]interface{}{
		"chat_id":    t.chatID,
		"text":       msgText,
		"parse_mode": "Markdown", // 注意如果因為特殊字元失敗可以拔掉 parse_mode 或處理脫逃字元
	}

	b, _ := json.Marshal(payload)
	resp, err := t.client.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram api error, status: %d", resp.StatusCode)
	}

	return nil
}
