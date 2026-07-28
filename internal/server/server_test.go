package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/pkg/logger"
	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/pkg/types"
	st "github.com/supercakecrumb/curly-notification-telegram-bot/internal/securetransformer"
)

const testSeed = "test_seed"

// newTestServer builds a Server with a queue of the given capacity and returns
// it alongside that queue, so tests can assert on what was actually published.
func newTestServer(t *testing.T, queueSize int) (*Server, chan types.NotificationRequest, *st.SecureTransformer) {
	t.Helper()

	transformer := st.NewSecureTransformer(testSeed)
	ch := make(chan types.NotificationRequest, queueSize)
	srv := NewServer(logger.New("debug"), transformer, ch)
	return srv, ch, transformer
}

// post sends body to the handler under test and returns the response.
func post(t *testing.T, h http.Handler, body any) *http.Response {
	t.Helper()

	var buf []byte
	switch v := body.(type) {
	case string:
		buf = []byte(v)
	default:
		var err error
		buf, err = json.Marshal(v)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
	}

	ts := httptest.NewServer(h)
	t.Cleanup(ts.Close)

	req, err := http.NewRequest(http.MethodPost, ts.URL, bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func TestSendNotification_Valid(t *testing.T) {
	srv, ch, transformer := newTestServer(t, 100)

	const telegramID = int64(12345)
	resp := post(t, http.HandlerFunc(srv.handleSendNotification), types.NotificationRequest{
		Text:       "Hello, world",
		TelegramID: strconv.FormatInt(telegramID, 10),
		Password:   transformer.Encode(telegramID),
	})

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
	if len(ch) != 1 {
		t.Fatalf("expected the notification to be queued, queue has %d", len(ch))
	}
	if got := (<-ch).Text; got != "Hello, world" {
		t.Errorf("queued text = %q, want %q", got, "Hello, world")
	}
}

func TestSendNotification_InvalidPassword(t *testing.T) {
	srv, ch, _ := newTestServer(t, 100)

	resp := post(t, http.HandlerFunc(srv.handleSendNotification), types.NotificationRequest{
		Text:       "Should fail",
		TelegramID: "12345",
		Password:   "invalid_password",
	})

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized, got %d", resp.StatusCode)
	}
	if len(ch) != 0 {
		t.Errorf("unauthenticated request was queued")
	}
}

func TestSendNotification_EmptyTelegramID(t *testing.T) {
	srv, _, _ := newTestServer(t, 100)

	resp := post(t, http.HandlerFunc(srv.handleSendNotification), types.NotificationRequest{
		Text:       "Hello, world",
		TelegramID: "",
		Password:   "somePassword",
	})

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for empty Telegram ID, got %d", resp.StatusCode)
	}
}

func TestSendNotification_EmptyPassword(t *testing.T) {
	srv, _, _ := newTestServer(t, 100)

	resp := post(t, http.HandlerFunc(srv.handleSendNotification), types.NotificationRequest{
		Text:       "Hello, world",
		TelegramID: "12345",
		Password:   "",
	})

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for empty password, got %d", resp.StatusCode)
	}
}

// TestSendNotification_EmptyText is the regression test for the bug this suite
// used to miss: the handler wrote a 400 for empty text but had no return, so
// the empty message was still queued and delivered. Asserting the status alone
// passed the whole time — the queue is the part that mattered.
func TestSendNotification_EmptyText(t *testing.T) {
	srv, ch, transformer := newTestServer(t, 100)

	resp := post(t, http.HandlerFunc(srv.handleSendNotification), types.NotificationRequest{
		Text:       "",
		TelegramID: "12345",
		Password:   transformer.Encode(12345),
	})

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for empty text, got %d", resp.StatusCode)
	}
	if len(ch) != 0 {
		t.Errorf("empty message was queued despite the 400; queue has %d", len(ch))
	}
}

func TestSendNotification_InvalidJSON(t *testing.T) {
	srv, _, _ := newTestServer(t, 100)

	resp := post(t, http.HandlerFunc(srv.handleSendNotification), "{not json")

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for malformed JSON, got %d", resp.StatusCode)
	}
}

func TestSendNotification_BodyTooLarge(t *testing.T) {
	srv, ch, transformer := newTestServer(t, 100)

	resp := post(t, http.HandlerFunc(srv.handleSendNotification), types.NotificationRequest{
		Text:       strings.Repeat("A", maxBodyBytes+1),
		TelegramID: "12345",
		Password:   transformer.Encode(12345),
	})

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected 413 for an oversized body, got %d", resp.StatusCode)
	}
	if len(ch) != 0 {
		t.Errorf("oversized request was queued")
	}
}

func TestSendNotification_InvalidFormat(t *testing.T) {
	srv, ch, transformer := newTestServer(t, 100)

	resp := post(t, http.HandlerFunc(srv.handleSendNotification), types.NotificationRequest{
		Text:       "hi",
		TelegramID: "12345",
		Password:   transformer.Encode(12345),
		Format:     "markdown",
	})

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 for an unknown format, got %d", resp.StatusCode)
	}
	if len(ch) != 0 {
		t.Errorf("request with an unknown format was queued")
	}
}

func TestSendNotification_FormatHTMLAccepted(t *testing.T) {
	srv, ch, transformer := newTestServer(t, 100)

	resp := post(t, http.HandlerFunc(srv.handleSendNotification), types.NotificationRequest{
		Text:       "<b>bold</b>",
		TelegramID: "12345",
		Password:   transformer.Encode(12345),
		Format:     types.FormatHTML,
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 for format=html, got %d", resp.StatusCode)
	}
	if got := (<-ch).Format; got != types.FormatHTML {
		t.Errorf("queued format = %q, want %q", got, types.FormatHTML)
	}
}

// TestSendNotification_QueueFull covers the load-shedding path: intake must
// return promptly rather than block forever when the sender has stalled.
func TestSendNotification_QueueFull(t *testing.T) {
	srv, ch, transformer := newTestServer(t, 1)

	// Occupy the only slot, and make the wait short so the test is quick.
	ch <- types.NotificationRequest{Text: "already queued", TelegramID: "12345"}
	original := enqueueTimeout
	enqueueTimeout = 50 * time.Millisecond
	t.Cleanup(func() { enqueueTimeout = original })

	start := time.Now()
	resp := post(t, http.HandlerFunc(srv.handleSendNotification), types.NotificationRequest{
		Text:       "should be shed",
		TelegramID: "12345",
		Password:   transformer.Encode(12345),
	})
	elapsed := time.Since(start)

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected 503 when the queue is full, got %d", resp.StatusCode)
	}
	if elapsed > time.Second {
		t.Errorf("intake blocked for %v; it should shed load promptly", elapsed)
	}
}

func TestRoutes_Healthz(t *testing.T) {
	srv, _, _ := newTestServer(t, 100)

	ts := httptest.NewServer(srv.Routes())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 from /healthz, got %d", resp.StatusCode)
	}
}

func TestRoutes_WrongMethod(t *testing.T) {
	srv, _, _ := newTestServer(t, 100)

	ts := httptest.NewServer(srv.Routes())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/send_notification")
	if err != nil {
		t.Fatalf("GET /send_notification: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405 for GET on the intake route, got %d", resp.StatusCode)
	}
}
