package server

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/pkg/types"
	st "github.com/supercakecrumb/curly-notification-telegram-bot/internal/securetransformer"
)

// Server wraps the HTTP logic and holds a reference to a SecureTransformer and a notification channel
type Server struct {
	logger           *slog.Logger
	transformer      *st.SecureTransformer
	NotificationChan chan types.NotificationRequest
}

// NewServer constructs a Server that uses the given transformer and has a notification channel
func NewServer(logger *slog.Logger, transformer *st.SecureTransformer, ch chan types.NotificationRequest) *Server {
	return &Server{
		logger:           logger,
		transformer:      transformer,
		NotificationChan: ch, // Buffered channel, capacity = 100
	}
}

// Start starts the HTTP server on the provided address (e.g., ":8080") and registers the routes.
func (s *Server) Start(addr string) {
	http.HandleFunc("/send_notification", s.handleSendNotification)

	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
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
