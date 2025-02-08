package telegram

import (
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/pkg/config"
)

type Bot struct {
	bot    *telego.Bot
	logger *slog.Logger
	config *config.Config
	bh     *th.BotHandler
}

func NewBot(token string, logger *slog.Logger, config *config.Config) (*Bot, error) {
	bot, err := telego.NewBot(config.TelegramToken)
	if err != nil {
		return nil, err
	}

	return &Bot{
		bot:    bot,
		logger: logger,
		config: config,
	}, nil
}

func (b *Bot) Start() {
	b.logger.Info("Starting bot...")

	// Notify admins about the shutdown
	b.NotifyAdmins("⚠️ The bot is starting.")

	// Use UpdatesViaLongPolling to handle updates
	updates, err := b.bot.UpdatesViaLongPolling(nil)
	if err != nil {
		b.logger.Error("Failed to start long polling", slog.String("error", err.Error()))
		return
	}

	// Create bot handler and specify from where to get updates
	b.bh, err = th.NewBotHandler(b.bot, updates)
	if err != nil {
		b.logger.Error("Failed to create new bot handler", slog.String("error", err.Error()))
		return
	}

	defer b.bh.Stop()
	defer b.bot.StopLongPolling()

	// Middleware in case of panic and no username
	b.bh.Use(
		th.PanicRecovery(),
	)

	// b.registerCommands()

	// b.registerAdminCommands()

	b.bh.Start()
}

func (b *Bot) Stop() {
	b.logger.Info("Stopping bot...")

	// Notify admins about the shutdown
	b.NotifyAdmins("⚠️ The bot is stopping. Please check the server for details.")

	// Stop the bot handler
	b.bh.Stop()
}

// NotifyAdmins sends a message to all admins
func (b *Bot) NotifyAdmins(message string) {
	_, err := b.bot.SendMessage(tu.Message(
		tu.ID(b.config.AdminTelgramID),
		message,
	))
	if err != nil {
		b.logger.Error("Failed to notify admin", slog.String("error", err.Error()))
	} else {
		b.logger.Info("Notified admin")
	}

}
