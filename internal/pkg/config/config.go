package config

import (
	"os"
)

type Config struct {
	TelegramToken  string
	LogLevel       string
	AdminTelgramID string
}

func LoadConfig() Config {
	return Config{
		TelegramToken:  os.Getenv("TELEGRAM_TOKEN"),
		LogLevel:       os.Getenv("LOG_LEVEL"),
		AdminTelgramID: os.Getenv("ADMIN_TELEGRAM_ID"),
	}
}
