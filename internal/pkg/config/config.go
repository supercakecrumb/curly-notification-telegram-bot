package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken   string
	LogLevel        string
	AdminTelgramID  int64
	TransformerSeed string
	ListenOn        string
	ApiDomain       string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		// It's not necessarily an error if .env doesn't exist;
		// you might want to log a warning instead. But let's treat it as an error here.
		fmt.Printf("Warning: Could not load .env file (err: %v)\n", err)
	}

	adminTgID, err := strconv.ParseInt(os.Getenv("ADMIN_TELEGRAM_ID"), 10, 64)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	return &Config{
		TelegramToken:   os.Getenv("TELEGRAM_TOKEN"),
		LogLevel:        os.Getenv("LOG_LEVEL"),
		TransformerSeed: os.Getenv("TRANSFORMER_SEED"),
		ListenOn:        os.Getenv("LISTEN_ON"),
		ApiDomain:       os.Getenv("API_DOMAIN"),
		AdminTelgramID:  adminTgID,
	}, nil
}
