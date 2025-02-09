package telegram

import (
	"bytes"
	"html/template"
	"log/slog"
	"strconv"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

type scriptData struct {
	TelegramID string
	Password   string
	APIDomain  string
}

// handleGetScript sends the Bash script with placeholders filled in.
func (b *Bot) handleGetScript(bot *telego.Bot, update telego.Update) {
	chatID := update.Message.Chat.ID
	username := update.Message.From.Username
	userTelegramID := tu.ID(chatID)
	hashedPassword := b.transformer.Encode(chatID)

	scriptData := scriptData{
		TelegramID: strconv.FormatInt(chatID, 10),
		Password:   hashedPassword,
		APIDomain:  b.apiDomain,
	}

	script, err := renderBashScript(scriptData)
	if err != nil {
		b.logger.Error("error rendering bash script", slog.String("error", err.Error()))
	}

	msg := telego.SendMessageParams{
		ChatID:    userTelegramID,
		Text:      script,
		ParseMode: telego.ModeHTML,
	}

	_, err = b.bot.SendMessage(&msg)
	if err != nil {
		b.logger.Error("error sending sciprt", slog.String("error", err.Error()), slog.String("username", username))
		return
	}
}

// RenderBashScript compiles and executes the template with the provided data.
func renderBashScript(data scriptData) (string, error) {
	// Parse the template string
	tmpl, err := template.New("bashScript").Parse(bashTemplate)
	if err != nil {
		return "", err
	}

	// Execute the template into a buffer
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
