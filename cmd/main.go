package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ichzzy/short-squeeze-detector/internal/binance"
	"github.com/ichzzy/short-squeeze-detector/internal/config"
	"github.com/ichzzy/short-squeeze-detector/internal/notifier"
	"github.com/ichzzy/short-squeeze-detector/internal/storage"
	"github.com/ichzzy/short-squeeze-detector/internal/strategy"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("config.Load failed: %v", err)
	}

	binanceClient := binance.NewClient(cfg.Binance.APIKey, cfg.Binance.APISecret)
	csvStorage := storage.NewCSVStorage("data")
	engine := strategy.NewEngine(cfg.Strategy.FundingRateThreshold, cfg.Strategy.OISurgeRatio)
	tgNotifier := notifier.NewTelegramNotifier(cfg.Telegram.BotToken, cfg.Telegram.ChatID)

	log.Printf("Starting short squeeze detector, monitoring %d symbols", len(cfg.Strategy.Symbols))
	log.Printf("Monitoring interval: %s", cfg.App.Interval.String())

	ticker := time.NewTicker(cfg.App.Interval)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	runJob(cfg.Strategy.Symbols, binanceClient, csvStorage, engine, tgNotifier)

	for {
		select {
		case <-ticker.C:
			runJob(cfg.Strategy.Symbols, binanceClient, csvStorage, engine, tgNotifier)
		case sig := <-sigChan:
			log.Printf("Received interrupt signal %v, shutting down...", sig)
			return
		}
	}
}

func runJob(
	symbols []string,
	client *binance.Client,
	csvStorage *storage.CSVStorage,
	engine *strategy.Engine,
	tgNotifier *notifier.TelegramNotifier,
) {
	log.Println("--- Running scheduled monitoring job ---")
	for _, symbol := range symbols {
		// 1. 抓取 binance 幣對數據
		data, err := client.FetchMarketData(symbol)
		if err != nil {
			log.Printf("[%s] Failed to fetch data: %v", symbol, err)
			continue
		}

		// 2. 寫入 CSV
		if err := csvStorage.Append(data); err != nil {
			log.Printf("[%s] Failed to write to CSV: %v", symbol, err)
		}

		// 3. 策略引擎運算
		alertEvent := engine.Process(data)

		// 4. 若有觸發告警，透過 Telegram 送出
		if alertEvent != nil {
			if err := tgNotifier.SendAlert(alertEvent); err != nil {
				log.Printf("[%s] Failed to send Telegram alert: %v", symbol, err)
			} else {
				log.Printf("[%s] Successfully sent Telegram alert", symbol)
			}
		}
	}
}
