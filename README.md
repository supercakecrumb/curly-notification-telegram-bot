# curly-notification-telegram-bot

A self-hosted notification relay: POST some JSON, get a Telegram message.
Anything that can run `curl` — a git hook, a cron job, CI, a long shell
command — can notify you **without holding a Telegram bot token**.

The only credential a caller needs is a per-chat password,
`base64url(HMAC-SHA256(seed, telegram_id))`. It is deterministic and
non-reversible, so it authorizes sending to exactly one chat and nothing else.
Message the bot `/getbashscript` to get a ready-made helper with yours filled in.

## API

### `POST /send_notification`

```json
{
  "text": "Backup finished",
  "telegram_id": "123456789",
  "password": "<per-chat password>",
  "format": "text"
}
```

| Field | Required | Notes |
|-------|----------|-------|
| `text` | yes | Rejected if empty. |
| `telegram_id` | yes | Numeric, as a string. |
| `password` | yes | Compared in constant time. |
| `format` | no | `text` (default) or `html`. |

Responses: `200` queued · `400` malformed/empty/unknown format · `401` bad
password · `413` body over 64 KiB · `503` queue full.

> **`format` matters more than it looks.** curly relays with Telegram's HTML
> parse mode, and delivery is **asynchronous** — the endpoint answers as soon as
> the password validates, then a worker calls `sendMessage`. If Telegram then
> rejects the text as malformed HTML, the message is dropped and only this
> service's log knows; the caller already got its `200`.
>
> So `format: "text"` (the default) means *curly escapes it for you* — send
> arbitrary commit subjects safely. Use `format: "html"` only when you are
> producing valid Telegram HTML yourself and want the markup preserved.

### `GET /healthz`

Returns `200 ok` whenever the process is alive. It deliberately does not probe
Telegram: a health check that failed during a Telegram outage would make the
orchestrator restart a service that is working fine. Notification delivery is
also decoupled from long polling, so the relay keeps working even if the
interactive half can't start.

## Configuration

Every variable below is **required** — startup fails with a list of what's
missing rather than half-working.

| Variable | Example | Purpose |
|---|---|---|
| `TELEGRAM_TOKEN` | `123456:AA…` | Bot token from @BotFather. |
| `TRANSFORMER_SEED` | `openssl rand -base64 32` | HMAC key the per-chat passwords derive from. **Rotating it invalidates every password already issued.** |
| `ADMIN_TELEGRAM_ID` | `123456789` | Receives startup/shutdown notices. |
| `API_DOMAIN` | `curly.example.com` | Baked into the `/getbashscript` helper. |

Optional: `LISTEN_ON` (default `:8080`), `LOG_LEVEL` (`debug`/`info`/`warn`/`error`, default `info`).

## Running

```bash
docker build -t curly .
docker run -d -p 8080:8080 --env-file .env curly
```

The image is multi-stage, distroless and runs as `nonroot`. `HEALTHCHECK` calls
the binary's own `--healthcheck` subcommand, which probes `/healthz`.

## Development

```bash
go test ./...
golangci-lint run ./...
```
