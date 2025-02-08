package types

// NotificationRequest represents the JSON payload we receive via POST /send_notification
type NotificationRequest struct {
	Text       string `json:"text"`
	TelegramID string `json:"telegram_id"`
	Password   string `json:"password"`
}
