package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/pkg/logger"
	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/pkg/types"
	st "github.com/supercakecrumb/curly-notification-telegram-bot/internal/securetransformer"
)

func TestSendNotification_Valid(t *testing.T) {
	// 1. Create the SecureTransformer
	seed := "test_seed"
	transformer := st.NewSecureTransformer(seed)

	// 2. Create our Server
	NotificationChan := make(chan types.NotificationRequest, 100)
	logger := logger.New("debug")
	srv := NewServer(logger, transformer, NotificationChan)

	// 3. Create an HTTP test server, pointing to our handler
	testServer := httptest.NewServer(http.HandlerFunc(srv.handleSendNotification))
	defer testServer.Close()

	// 4. Build a valid request (correct password)
	telegramID := int64(12345)
	validPassword := transformer.Encode(telegramID)

	body := types.NotificationRequest{
		Text:       "Hello, world",
		TelegramID: strconv.FormatInt(telegramID, 10),
		Password:   validPassword,
	}

	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, testServer.URL, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	// 5. Perform the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error making valid request: %v", err)
	}
	defer resp.Body.Close()

	// 6. Validate
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestSendNotification_InvalidPassword(t *testing.T) {
	transformer := st.NewSecureTransformer("test_seed")
	NotificationChan := make(chan types.NotificationRequest, 100)
	logger := logger.New("debug")
	srv := NewServer(logger, transformer, NotificationChan)

	testServer := httptest.NewServer(http.HandlerFunc(srv.handleSendNotification))
	defer testServer.Close()

	// Wrong password
	body := types.NotificationRequest{
		Text:       "Should fail",
		TelegramID: "12345",
		Password:   "invalid_password",
	}

	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, testServer.URL, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error making invalid request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized, got %d", resp.StatusCode)
	}
}

func TestSendNotification_EmptyTelegramID(t *testing.T) {
	transformer := st.NewSecureTransformer("test_seed")
	NotificationChan := make(chan types.NotificationRequest, 100)
	logger := logger.New("debug")
	srv := NewServer(logger, transformer, NotificationChan)

	testServer := httptest.NewServer(http.HandlerFunc(srv.handleSendNotification))
	defer testServer.Close()

	body := types.NotificationRequest{
		Text:       "Hello, world",
		TelegramID: "", // Empty TG ID
		Password:   "somePassword",
	}

	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, testServer.URL, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error making request with empty TG ID: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for empty Telegram ID, got %d", resp.StatusCode)
	}
}

func TestSendNotification_EmptyPassword(t *testing.T) {
	transformer := st.NewSecureTransformer("test_seed")
	NotificationChan := make(chan types.NotificationRequest, 100)
	logger := logger.New("debug")
	srv := NewServer(logger, transformer, NotificationChan)

	testServer := httptest.NewServer(http.HandlerFunc(srv.handleSendNotification))
	defer testServer.Close()

	body := types.NotificationRequest{
		Text:       "Hello, world",
		TelegramID: "12345",
		Password:   "", // Empty password
	}

	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, testServer.URL, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error making request with empty password: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for empty password, got %d", resp.StatusCode)
	}
}

func TestSendNotification_EmptyText(t *testing.T) {
	transformer := st.NewSecureTransformer("test_seed")
	NotificationChan := make(chan types.NotificationRequest, 100)
	logger := logger.New("debug")
	srv := NewServer(logger, transformer, NotificationChan)

	testServer := httptest.NewServer(http.HandlerFunc(srv.handleSendNotification))
	defer testServer.Close()

	// We'll assume empty text is allowed or not. If you want to reject it, you'll enforce it in your code.
	body := types.NotificationRequest{
		Text:       "", // Empty text
		TelegramID: "12345",
		Password:   transformer.Encode(12345),
	}

	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, testServer.URL, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error making request with empty text: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for empty Telegram ID, got %d", resp.StatusCode)
	}
}
