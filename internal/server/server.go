// Package server exposes curly's HTTP surface: the notification intake and a
// health probe.
package server

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/pkg/types"
	st "github.com/supercakecrumb/curly-notification-telegram-bot/internal/securetransformer"
)

const (
	// maxBodyBytes caps a request body. A notification is a short line of
	// text; without a cap, json.Decode would read whatever it is handed.
	maxBodyBytes = 64 << 10 // 64 KiB

	// readHeaderTimeout bounds how long a client may take to send its
	// headers, which is what makes Slowloris cheap to defend against.
	readHeaderTimeout = 10 * time.Second

	// shutdownTimeout bounds graceful shutdown of in-flight requests.
	shutdownTimeout = 10 * time.Second
)

// enqueueTimeout is how long intake waits for room in the notification queue
// before shedding the request. It exists so a stalled Telegram sender produces
// a fast 503 instead of an HTTP handler that blocks forever holding a
// connection open. A variable rather than a constant so tests can shorten it.
var enqueueTimeout = 2 * time.Second

// Server wraps the HTTP logic and holds a reference to a SecureTransformer and
// a notification channel.
type Server struct {
	logger           *slog.Logger
	transformer      *st.SecureTransformer
	NotificationChan chan types.NotificationRequest
	server           *http.Server
}

// NewServer constructs a Server that uses the given transformer and publishes
// accepted notifications to ch.
func NewServer(logger *slog.Logger, transformer *st.SecureTransformer, ch chan types.NotificationRequest) *Server {
	return &Server{
		logger:           logger,
		transformer:      transformer,
		NotificationChan: ch,
	}
}

// Routes returns the server's handler. Exported so tests can exercise the real
// routing table rather than a hand-wired handler.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /send_notification", s.handleSendNotification)
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	return mux
}

// Start serves until Stop is called. It returns nil on a graceful shutdown and
// an error if the listener could not be established — the caller decides what
// that means, rather than the old log.Fatalf which killed the process from
// inside a goroutine and skipped every deferred cleanup.
func (s *Server) Start(addr string) error {
	s.server = &http.Server{
		Addr:              addr,
		Handler:           s.Routes(),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	s.logger.Info("server listening", slog.String("addr", addr))
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}
	return nil
}

// Stop shuts the server down gracefully, waiting for in-flight requests. Once
// it returns, no handler is still running — which is what makes it safe for the
// caller to then close the notification channel.
func (s *Server) Stop() {
	if s.server == nil {
		return // never started
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	s.logger.Info("shutting down server")
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("server shutdown", slog.String("error", err.Error()))
		return
	}
	s.logger.Info("server stopped gracefully")
}

// handleHealthz reports process liveness. It deliberately checks nothing
// external: a health probe that failed whenever Telegram was unreachable would
// make the orchestrator restart a service that is working fine.
func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, "ok")
}

// handleSendNotification decodes JSON, validates the per-chat password, then
// publishes the notification for the Telegram sender to deliver.
func (s *Server) handleSendNotification(w http.ResponseWriter, r *http.Request) {
	// Routes() constrains the method, but the handler is also mounted directly
	// in tests, so keep the check here too.
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req types.NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	telegramID, err := strconv.ParseInt(req.TelegramID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid telegram_id: must be numeric", http.StatusBadRequest)
		return
	}

	// Constant-time comparison: this is a credential check, so it must not
	// leak how much of the password matched via response timing.
	expected := s.transformer.Encode(telegramID)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(req.Password)) != 1 {
		http.Error(w, "Unauthorized: invalid password", http.StatusUnauthorized)
		return
	}

	// This guard used to write a 400 and then fall through, so an empty
	// message was still queued and delivered.
	if req.Text == "" {
		http.Error(w, "Message is empty. Nothing to send.", http.StatusBadRequest)
		return
	}

	// Reject an unknown format rather than silently picking a default: the
	// difference between the two decides whether the caller's markup is
	// escaped, and guessing wrong is exactly the failure that loses messages.
	switch req.Format {
	case "", types.FormatText, types.FormatHTML:
	default:
		http.Error(w, `Invalid format: must be "text" or "html"`, http.StatusBadRequest)
		return
	}

	timer := time.NewTimer(enqueueTimeout)
	defer timer.Stop()

	select {
	case s.NotificationChan <- req:
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "Notification queued")
	case <-r.Context().Done():
		// The client hung up; nothing useful left to write.
		s.logger.Warn("intake abandoned by client", slog.String("chat_id", req.TelegramID))
	case <-timer.C:
		// The queue is full, meaning the sender is stalled or Telegram is
		// rate-limiting us. Shed load loudly instead of blocking.
		s.logger.Error("notification queue full; shedding request",
			slog.String("chat_id", req.TelegramID))
		http.Error(w, "Notification queue full, try again later", http.StatusServiceUnavailable)
	}
}
