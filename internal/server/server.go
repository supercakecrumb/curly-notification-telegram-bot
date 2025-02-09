package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/pkg/types"
	st "github.com/supercakecrumb/curly-notification-telegram-bot/internal/securetransformer"
)

// Server wraps the HTTP logic and holds a reference to a SecureTransformer and a notification channel
type Server struct {
	logger           *slog.Logger
	transformer      *st.SecureTransformer
	NotificationChan chan types.NotificationRequest
	server           *http.Server
}

// NewServer constructs a Server that uses the given transformer and has a notification channel
func NewServer(logger *slog.Logger, transformer *st.SecureTransformer, ch chan types.NotificationRequest) *Server {
	return &Server{
		logger:           logger,
		transformer:      transformer,
		NotificationChan: ch, // Buffered channel, capacity = 100
	}
}

// Start registers routes, creates an *http.Server, and serves until Stop() is called.
func (s *Server) Start(addr string) {
	mux := http.NewServeMux()

	// Register your routes
	mux.HandleFunc("/send_notification", s.handleSendNotification)

	s.server = &http.Server{
		Addr:    addr,
		Handler: mux, // Use our mux with the handler
	}

	log.Printf("Server listening on %s", addr)
	// ListenAndServe will block until the server is stopped or fails
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}

// Stop shuts down the server gracefully with a 1 second timeout.
func (s *Server) Stop() {
	if s.server == nil {
		return // Server never started or already stopped
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	log.Println("Shutting down server...")

	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf("Server Shutdown error: %v", err)
	} else {
		log.Println("Server stopped gracefully.")
	}
}

// handleSendNotification decodes JSON, validates password, then pushes to NotificationChan.
func (s *Server) handleSendNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req types.NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Parse telegram_id (string) to int64
	telegramID, err := strconv.ParseInt(req.TelegramID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid telegram_id: must be numeric", http.StatusBadRequest)
		return
	}

	// Validate the provided password: must match transformer.Encode(telegramID)
	expectedPassword := s.transformer.Encode(telegramID)
	if expectedPassword != req.Password {
		http.Error(w, "Unauthorized: invalid password", http.StatusUnauthorized)
		return
	}

	if req.Text == "" {
		http.Error(w, "Message is empty. Nothing to send.", http.StatusBadRequest)
	}

	// If we reach here, the request is valid → push the notification
	s.NotificationChan <- req

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Notification queued")
}
