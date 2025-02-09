package telegram

import (
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/pkg/types"
	st "github.com/supercakecrumb/curly-notification-telegram-bot/internal/securetransformer"
)

type Bot struct {
	bot              *telego.Bot
	logger           *slog.Logger
	adminID          int64
	apiDomain        string
	bh               *th.BotHandler
	st               *st.SecureTransformer
	notificationChan chan types.NotificationRequest
	transformer      *st.SecureTransformer
}

func NewBot(logger *slog.Logger, token, apiDomain string, adminID int64, transformer *st.SecureTransformer, ch chan types.NotificationRequest) (*Bot, error) {
	bot, err := telego.NewBot(token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		bot:              bot,
		logger:           logger,
		adminID:          adminID,
		notificationChan: ch,
		transformer:      transformer,
		apiDomain:        apiDomain,
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

	b.registerCommands()

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
		tu.ID(b.adminID),
		message,
	))
	if err != nil {
		b.logger.Error("Failed to notify admin", slog.String("error", err.Error()))
	} else {
		b.logger.Info("Notified admin")
	}

}
