package main

import (
	"fmt"

	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/pkg/config"
	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/pkg/logger"
	st "github.com/supercakecrumb/curly-notification-telegram-bot/internal/securetransformer"
)

func main() {
	config, err := config.LoadConfig()

	if err != nil {
		fmt.Printf("error loading config: %v", err.Error())
	}

	logger := logger.New(config.LogLevel)

	secureTransformer := st.NewSecureTransformer(config.TransformerSeed)
}
