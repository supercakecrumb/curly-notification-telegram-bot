package config

import (
	"strings"
	"testing"
)

// setValid populates every required variable, so an individual test can then
// clear exactly the one it cares about.
func setValid(t *testing.T) {
	t.Helper()
	t.Setenv("TELEGRAM_TOKEN", "123456:test-token")
	t.Setenv("TRANSFORMER_SEED", strings.Repeat("s", 44))
	t.Setenv("API_DOMAIN", "curly.example.com")
	t.Setenv("ADMIN_TELEGRAM_ID", "12345")
	t.Setenv("LISTEN_ON", "")
	t.Setenv("LOG_LEVEL", "")
}

func TestLoadConfig_Valid(t *testing.T) {
	setValid(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}
	if cfg.AdminTelegramID != 12345 {
		t.Errorf("AdminTelegramID = %d, want 12345", cfg.AdminTelegramID)
	}
	if cfg.ListenOn != defaultListenOn {
		t.Errorf("ListenOn = %q, want the default %q", cfg.ListenOn, defaultListenOn)
	}
}

// TestLoadConfig_MissingRequired is the fail-closed guarantee. An empty
// TRANSFORMER_SEED in particular used to sail through here and panic later in
// securetransformer; a missing ADMIN_TELEGRAM_ID returned an error that main
// printed and ignored, then dereferenced the nil *Config.
func TestLoadConfig_MissingRequired(t *testing.T) {
	for _, missing := range []string{
		"TELEGRAM_TOKEN",
		"TRANSFORMER_SEED",
		"API_DOMAIN",
		"ADMIN_TELEGRAM_ID",
	} {
		t.Run(missing, func(t *testing.T) {
			setValid(t)
			t.Setenv(missing, "")

			cfg, err := LoadConfig()
			if err == nil {
				t.Fatalf("LoadConfig() with %s unset returned nil error", missing)
			}
			if cfg != nil {
				t.Errorf("LoadConfig() returned a non-nil config alongside an error")
			}
			if !strings.Contains(err.Error(), missing) {
				t.Errorf("error %q does not name the missing variable %s", err, missing)
			}
		})
	}
}

// All missing variables should be reported together, so a fresh deploy is fixed
// in one pass instead of one restart per variable.
func TestLoadConfig_ReportsAllMissingAtOnce(t *testing.T) {
	setValid(t)
	t.Setenv("TELEGRAM_TOKEN", "")
	t.Setenv("API_DOMAIN", "")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("LoadConfig() returned nil error")
	}
	for _, want := range []string{"TELEGRAM_TOKEN", "API_DOMAIN"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %s", err, want)
		}
	}
}

func TestLoadConfig_NonNumericAdminID(t *testing.T) {
	setValid(t)
	t.Setenv("ADMIN_TELEGRAM_ID", "not-a-number")

	if _, err := LoadConfig(); err == nil {
		t.Fatal("LoadConfig() accepted a non-numeric ADMIN_TELEGRAM_ID")
	}
}
