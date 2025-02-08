package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	TelegramToken   string
	LogLevel        string
	AdminTelgramID  int64
	TransformerSeed string
}

func LoadConfig() (*Config, error) {
	adminTgID, err := strconv.ParseInt(os.Getenv("ADMIN_TELEGRAM_ID"), 10, 64)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	return &Config{
		TelegramToken:   os.Getenv("TELEGRAM_TOKEN"),
		LogLevel:        os.Getenv("LOG_LEVEL"),
		TransformerSeed: os.Getenv("TRANSFORMER_SEED"),
		AdminTelgramID:  adminTgID,
	}, nil
}
