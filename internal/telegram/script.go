package telegram

import (
	"fmt"
	"log/slog"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

// handleGetScript sends the Bash script with placeholders filled in.
func (b *Bot) handleGetScript(bot *telego.Bot, update telego.Update) {
	chatID := update.Message.Chat.ID
	username := update.Message.From.Username
	userTelegramID := tu.ID(chatID)
	hashedPassword := b.transformer.Encode(chatID)

	script := fmt.Sprintf(bashScript, userTelegramID, hashedPassword, b.apiDomain)

	msg := telego.SendMessageParams{
		ChatID:    userTelegramID,
		Text:      script,
		ParseMode: telego.ModeHTML,
	}

	_, err := b.bot.SendMessage(&msg)
	if err != nil {
		b.logger.Error("error sending sciprt", slog.String("error", err.Error()), slog.String("username", username))
		return
	}
}
