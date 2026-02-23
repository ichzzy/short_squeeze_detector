package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Binance  BinanceConfig  `mapstructure:"binance"`
	Strategy StrategyConfig `mapstructure:"strategy"`
	Telegram TelegramConfig `mapstructure:"telegram"`
}

type AppConfig struct {
	Interval time.Duration `mapstructure:"interval"`
}

type BinanceConfig struct {
	APIKey    string `mapstructure:"api_key"`
	APISecret string `mapstructure:"api_secret"`
}

type StrategyConfig struct {
	Symbols              []string `mapstructure:"symbols"`
	FundingRateThreshold float64  `mapstructure:"funding_rate_threshold"`
	OISurgeRatio         float64  `mapstructure:"oi_surge_ratio"`
}

type TelegramConfig struct {
	BotToken string `mapstructure:"bot_token"`
	ChatID   string `mapstructure:"chat_id"`
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		// Ensure code works if workdir is under ./cmd
		viper.SetConfigFile("../" + path)
		if err2 := viper.ReadInConfig(); err2 != nil {
			return nil, fmt.Errorf("failed to read config file: %v", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %v", err)
	}

	return &cfg, nil
}
