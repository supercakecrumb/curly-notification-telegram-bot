// Package config loads and validates curly's environment configuration.
//
// Loading FAILS CLOSED: everything the service cannot work without is checked
// here and reported together, so a misconfigured deploy dies immediately with a
// list of what's missing instead of failing later somewhere unrelated. An empty
// TRANSFORMER_SEED used to reach securetransformer and panic there; an
// unparseable ADMIN_TELEGRAM_ID returned an error the caller ignored, leaving a
// nil *Config to be dereferenced on the next line.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// defaultListenOn is used when LISTEN_ON is unset.
const defaultListenOn = ":8080"

// weakSeedLen is the length below which TRANSFORMER_SEED is reported as weak.
// This warns rather than blocks: an existing deployment's seed cannot be
// rotated without invalidating every password already issued from it.
const weakSeedLen = 32

// Config holds curly's runtime configuration.
type Config struct {
	TelegramToken   string
	LogLevel        string
	AdminTelegramID int64
	TransformerSeed string
	ListenOn        string
	APIDomain       string
}

// LoadConfig reads configuration from the environment, optionally seeded by a
// .env file, and validates it. All missing values are reported in one error.
func LoadConfig() (*Config, error) {
	// A missing .env is normal in production, where the platform injects real
	// environment variables. Only a malformed one is worth mentioning.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		slog.Debug("config: no .env loaded", "error", err)
	}

	cfg := &Config{
		TelegramToken:   os.Getenv("TELEGRAM_TOKEN"),
		LogLevel:        os.Getenv("LOG_LEVEL"),
		TransformerSeed: os.Getenv("TRANSFORMER_SEED"),
		ListenOn:        os.Getenv("LISTEN_ON"),
		APIDomain:       os.Getenv("API_DOMAIN"),
	}
	if cfg.ListenOn == "" {
		cfg.ListenOn = defaultListenOn
	}

	var missing []string
	if cfg.TelegramToken == "" {
		missing = append(missing, "TELEGRAM_TOKEN")
	}
	if cfg.TransformerSeed == "" {
		// Without a seed, every password would be derived from an empty HMAC
		// key — computable by anyone who knows the scheme.
		missing = append(missing, "TRANSFORMER_SEED")
	}
	if cfg.APIDomain == "" {
		// /getbashscript renders this into the helper it hands out.
		missing = append(missing, "API_DOMAIN")
	}

	if adminRaw := os.Getenv("ADMIN_TELEGRAM_ID"); adminRaw == "" {
		missing = append(missing, "ADMIN_TELEGRAM_ID")
	} else {
		id, err := strconv.ParseInt(adminRaw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ADMIN_TELEGRAM_ID must be a numeric Telegram id, got %q", adminRaw)
		}
		cfg.AdminTelegramID = id
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	if len(cfg.TransformerSeed) < weakSeedLen {
		slog.Warn("config: TRANSFORMER_SEED is short; generate one with `openssl rand -base64 32`",
			"length", len(cfg.TransformerSeed))
	}

	return cfg, nil
}
