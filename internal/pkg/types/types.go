package types

// Message formats accepted in NotificationRequest.Format.
const (
	// FormatText means Text is plain text and curly escapes it before
	// handing it to Telegram. This is the default for a reason: curly relays
	// with ParseMode=HTML, so an unescaped "<" or "&" makes Telegram reject
	// the message — and because delivery is asynchronous, the caller already
	// got its 200 and never learns the message was dropped.
	FormatText = "text"

	// FormatHTML means the caller has produced valid Telegram HTML itself and
	// takes responsibility for it. Used by callers that want real markup.
	FormatHTML = "html"
)

// NotificationRequest represents the JSON payload we receive via POST /send_notification
type NotificationRequest struct {
	Text       string `json:"text"`
	TelegramID string `json:"telegram_id"`
	Password   string `json:"password"`

	// Format is FormatText (the default when empty) or FormatHTML.
	Format string `json:"format,omitempty"`
}
