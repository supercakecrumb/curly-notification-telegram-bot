package telegram

import (
	"testing"

	"github.com/supercakecrumb/curly-notification-telegram-bot/internal/pkg/types"
)

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain text untouched", "deploy finished", "deploy finished"},
		{"less than", "a < b", "a &lt; b"},
		{"greater than", "a > b", "a &gt; b"},
		{"ampersand", "R&D", "R&amp;D"},
		{
			// The reason this matters: Go commit subjects are full of these,
			// and an unescaped one made Telegram reject the message while the
			// caller saw a 200.
			"go commit subject",
			"fix: guard a < b && c > d in <-chan Foo[T]",
			"fix: guard a &lt; b &amp;&amp; c &gt; d in &lt;-chan Foo[T]",
		},
		{
			// Single-pass replacement: the & introduced by escaping "<" must
			// not itself be escaped again into "&amp;lt;".
			"no double escaping",
			"<b>",
			"&lt;b&gt;",
		},
		{
			// Quotes are deliberately left alone — html.EscapeString would
			// turn them into &#34;, which Telegram renders literally.
			"quotes preserved",
			`say "hi" it's fine`,
			`say "hi" it's fine`,
		},
		{"already escaped text is escaped again", "&amp;", "&amp;amp;"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EscapeHTML(tt.in); got != tt.want {
				t.Errorf("EscapeHTML(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRenderText(t *testing.T) {
	const raw = "<b>bold</b> & <i>italic</i>"

	tests := []struct {
		name   string
		format string
		want   string
	}{
		{
			"empty format defaults to escaping",
			"",
			"&lt;b&gt;bold&lt;/b&gt; &amp; &lt;i&gt;italic&lt;/i&gt;",
		},
		{
			"explicit text format escapes",
			types.FormatText,
			"&lt;b&gt;bold&lt;/b&gt; &amp; &lt;i&gt;italic&lt;/i&gt;",
		},
		{
			"html format passes through untouched",
			types.FormatHTML,
			raw,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderText(types.NotificationRequest{Text: raw, Format: tt.format})
			if got != tt.want {
				t.Errorf("renderText(format=%q) = %q, want %q", tt.format, got, tt.want)
			}
		})
	}
}
