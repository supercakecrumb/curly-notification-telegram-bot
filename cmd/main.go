package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/pkg/config"
	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/pkg/logger"
	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/pkg/types"
	st "github.com/supercakecrumb/curly-notification-telegram-bot/internal/securetransformer"
	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/server"
	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/telegram"
)

func main() {
	config, err := config.LoadConfig()

	if err != nil {
		fmt.Printf("error loading config: %v", err.Error())
	}

	logger := logger.New(config.LogLevel)

	secureTransformer := st.NewSecureTransformer(config.TransformerSeed)

	notificationChan := make(chan types.NotificationRequest, 100)

	bot, err := telegram.NewBot(logger, config.TelegramToken, config.ApiDomain, config.AdminTelgramID, secureTransformer, notificationChan)
	if err != nil {
		logger.Error("error initializing bot", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Create a context that is cancelled on OS interrupt or terminate signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start the bot in a separate goroutine
	go func() {
		logger.Info("Bot is starting...")
		bot.Start()
	}()

	server := server.NewServer(logger, secureTransformer, notificationChan)

	go func() {
		logger.Info("Server is starting...")
		server.Start(config.ApiDomain) // blocks in this goroutine
	}()

	// Wait for the context to be cancelled (signal received)
	<-ctx.Done()
	logger.Info("Shutting down gracefully...")

	// Stop the bot handler
	bot.Stop()
	server.Stop()
}
