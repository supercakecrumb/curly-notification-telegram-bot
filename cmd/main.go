// Command curly is a self-hosted notification relay: an HTTP receiver turns a
// POST into a Telegram message, so anything that can run curl can notify you
// without holding a Telegram bot token.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/pkg/config"
	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/pkg/logger"
	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/pkg/types"
	st "github.com/supercakecrumb/curly-notification-telegram-bot/internal/securetransformer"
	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/server"
	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/telegram"
)

const (
	// notifyQueueSize is how many accepted notifications may wait for the
	// Telegram sender. Beyond this the HTTP layer sheds load with a 503.
	notifyQueueSize = 100

	// senderDrainTimeout bounds how long shutdown waits for queued
	// notifications to actually reach Telegram.
	senderDrainTimeout = 10 * time.Second

	// defaultListenOn mirrors the config default, for the healthcheck probe.
	defaultListenOn = ":8080"
)

func main() {
	// Docker's HEALTHCHECK runs the same binary with this flag.
	if len(os.Args) > 1 && (os.Args[1] == "--healthcheck" || os.Args[1] == "healthcheck") {
		os.Exit(runHealthcheck())
	}

	if err := run(); err != nil {
		slog.Error("curly: fatal", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// runHealthcheck probes the local /healthz endpoint. It reads LISTEN_ON
// directly from the environment rather than going through config.LoadConfig, so
// the probe stays cheap and does not fail merely because an unrelated variable
// is missing.
func runHealthcheck() int {
	client := &http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get("http://" + healthcheckAddr() + "/healthz")
	if err != nil {
		return 1
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}

// healthcheckAddr turns the listen address into something dialable from inside
// the container: a wildcard host is not a valid destination.
func healthcheckAddr() string {
	listen := os.Getenv("LISTEN_ON")
	if listen == "" {
		listen = defaultListenOn
	}

	host, port, err := net.SplitHostPort(listen)
	if err != nil {
		return "127.0.0.1:8080"
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, port)
}

func run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.LogLevel)
	slog.SetDefault(log)

	transformer := st.NewSecureTransformer(cfg.TransformerSeed)
	notificationChan := make(chan types.NotificationRequest, notifyQueueSize)

	bot, err := telegram.NewBot(log, cfg.TelegramToken, cfg.APIDomain, cfg.AdminTelegramID, transformer, notificationChan)
	if err != nil {
		return fmt.Errorf("init bot: %w", err)
	}
	srv := server.NewServer(log, transformer, notificationChan)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// The relay is started independently of long polling: delivering
	// notifications needs only the API client, so it keeps working even if the
	// interactive half fails to start.
	bot.StartNotificationListener()

	errCh := make(chan error, 2)
	go func() {
		if err := bot.Start(); err != nil {
			errCh <- fmt.Errorf("bot: %w", err)
		}
	}()
	go func() {
		if err := srv.Start(cfg.ListenOn); err != nil {
			errCh <- fmt.Errorf("server: %w", err)
		}
	}()

	var runErr error
	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case runErr = <-errCh:
		log.Error("component failed, shutting down", slog.String("error", runErr.Error()))
	}

	// Shutdown order is load-bearing. Stopping the HTTP server first guarantees
	// no handler is still running, so nothing can publish to the channel; only
	// then is closing it safe. Closing it makes the sender drain what is
	// already queued rather than lose it, and WaitSender gives that drain a
	// bounded window before the bot itself goes away.
	srv.Stop()
	close(notificationChan)
	bot.WaitSender(senderDrainTimeout)
	bot.Stop()

	return runErr
}
