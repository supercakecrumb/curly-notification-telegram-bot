package telegram

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func (b *Bot) registerCommands() {
	// Register command handlers
	b.bh.Handle(b.handleStart, th.CommandEqual("start"))
	b.bh.Handle(b.handleHelp, th.CommandEqual("help"))

	b.bh.Handle(b.handleGetScript, th.CommandEqual("getbashscript"))
}

// Handle /start command
func (b *Bot) handleStart(bot *telego.Bot, update telego.Update) {
	chatID := update.Message.Chat.ID

	welcomeMessage := startText

	msg := telego.SendMessageParams{
		ChatID:    tu.ID(chatID),
		Text:      welcomeMessage,
		ParseMode: telego.ModeHTML,
	}

	_, err := bot.SendMessage(&msg)
	if err != nil {
		b.logger.Error("Failed to send start message", "error", err)
	}
}

// Handle /help command
func (b *Bot) handleHelp(bot *telego.Bot, update telego.Update) {
	chatID := update.Message.Chat.ID

	helpMessage := helpText

	msg := telego.SendMessageParams{
		ChatID:    tu.ID(chatID),
		Text:      helpMessage,
		ParseMode: telego.ModeHTML,
	}

	_, err := bot.SendMessage(&msg)
	if err != nil {
		b.logger.Error("Failed to send start message", "error", err)
	}
}
