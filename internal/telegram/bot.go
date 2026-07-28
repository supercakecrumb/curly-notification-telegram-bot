package telegram

import (
	"fmt"
	"log/slog"
	"time"

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
	notificationChan chan types.NotificationRequest
	transformer      *st.SecureTransformer
	senderDone       chan struct{}
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
		senderDone:       make(chan struct{}),
	}, nil
}

// Start runs the interactive command handler and blocks until Stop is called.
//
// It deliberately does NOT own the notification sender: relaying a notification
// only needs the Telegram API client, not long polling, so the caller starts
// StartNotificationListener separately. That way a long-polling failure — a
// second instance holding the same bot token, say — costs us /start and /help
// but still delivers every notification.
func (b *Bot) Start() error {
	b.logger.Info("starting bot")
	b.NotifyAdmins("⚠️ The bot is starting.")

	updates, err := b.bot.UpdatesViaLongPolling(nil)
	if err != nil {
		return fmt.Errorf("start long polling: %w", err)
	}
	defer b.bot.StopLongPolling()

	b.bh, err = th.NewBotHandler(b.bot, updates)
	if err != nil {
		return fmt.Errorf("create bot handler: %w", err)
	}
	defer b.bh.Stop()

	// Middleware in case of panic
	b.bh.Use(th.PanicRecovery())

	b.registerCommands()

	b.bh.Start()
	return nil
}

// Stop halts the command handler. It is safe to call even when Start failed
// before the handler existed — b.bh is nil in that case, and calling Stop on it
// used to panic during shutdown.
func (b *Bot) Stop() {
	b.logger.Info("stopping bot")
	b.NotifyAdmins("⚠️ The bot is stopping. Please check the server for details.")

	if b.bh != nil {
		b.bh.Stop()
	}
}

// WaitSender blocks until the notification sender has drained its channel, or
// until timeout elapses. Call it after closing the notification channel so
// queued messages still get delivered during a graceful shutdown.
func (b *Bot) WaitSender(timeout time.Duration) {
	select {
	case <-b.senderDone:
		b.logger.Info("notification sender drained")
	case <-time.After(timeout):
		b.logger.Warn("timed out waiting for notification sender to drain",
			slog.Duration("timeout", timeout))
	}
}

// NotifyAdmins sends a message to the admin chat.
func (b *Bot) NotifyAdmins(message string) {
	_, err := b.bot.SendMessage(tu.Message(
		tu.ID(b.adminID),
		message,
	))
	if err != nil {
		b.logger.Error("failed to notify admin", slog.String("error", err.Error()))
		return
	}
	b.logger.Info("notified admin")
}
