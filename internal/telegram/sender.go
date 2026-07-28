package telegram

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/pkg/types"
)

// htmlEscaper escapes the three characters Telegram's HTML parse mode treats as
// markup. It is a single-pass replacer, so an ampersand produced by escaping is
// never escaped again — the double-escaping trap you get from running three
// sequential strings.ReplaceAll in the wrong order.
//
// Only these three are escaped. Quotes are left alone deliberately:
// html.EscapeString would render them as the numeric entities &#34;/&#39;,
// which Telegram does not decode, so they would show up literally in the chat.
var htmlEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
)

// EscapeHTML makes arbitrary text safe to send with ParseMode=HTML.
func EscapeHTML(s string) string {
	return htmlEscaper.Replace(s)
}

// renderText applies the request's format policy, returning what should be sent
// to Telegram.
func renderText(note types.NotificationRequest) string {
	if note.Format == types.FormatHTML {
		return note.Text
	}
	return EscapeHTML(note.Text)
}

// StartNotificationListener drains the notification channel and delivers each
// message to Telegram. It returns immediately; the caller can wait for the
// drain to finish with WaitSender.
func (b *Bot) StartNotificationListener() {
	go func() {
		defer close(b.senderDone)

		for note := range b.notificationChan {
			// The HTTP layer validates this before queueing, so a failure
			// here means the message reached the channel by another route.
			// It used to be assigned and then silently overwritten by the
			// SendMessage error, which would have sent to chat id 0.
			chatID, err := strconv.ParseInt(note.TelegramID, 10, 64)
			if err != nil {
				b.logger.Error("dropping notification with non-numeric chat id",
					slog.String("chat_id", note.TelegramID),
					slog.String("error", err.Error()))
				continue
			}

			message := telego.SendMessageParams{
				ChatID:    tu.ID(chatID),
				Text:      renderText(note),
				ParseMode: telego.ModeHTML,
			}

			if _, err := b.bot.SendMessage(&message); err != nil {
				// Delivery is asynchronous, so nobody is waiting on this
				// error — the caller was told "queued" long ago. This log is
				// the only record that a notification was lost.
				b.logger.Error("failed to send Telegram message",
					slog.String("chat_id", note.TelegramID),
					slog.String("error", err.Error()))
				continue
			}

			b.logger.Info("message sent",
				slog.String("chat_id", note.TelegramID),
				slog.Int("text_len", len(note.Text)))
		}
	}()
}
