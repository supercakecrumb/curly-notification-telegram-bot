package telegram

import (
	"log/slog"
	"strconv"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

func (b *Bot) StartNotificationListener() {
	go func() {
		for note := range b.notificationChan {

			chatID, err := strconv.ParseInt(note.TelegramID, 10, 64)

			message := telego.SendMessageParams{
				ChatID:    tu.ID(chatID),
				Text:      note.Text,
				ParseMode: telego.ModeHTML,
			}

			_, err = b.bot.SendMessage(&message)
			if err != nil {
				b.logger.Error("Failed to send Telegram message",
					slog.String("chat_id", note.TelegramID),
					slog.String("error", err.Error()))
			} else {
				b.logger.Info("Message sent successfully",
					slog.String("chat_id", note.TelegramID),
					slog.String("text", note.Text))
			}
		}
	}()
}
