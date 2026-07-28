# Changelog

Release notes are read verbatim from the section matching the pushed tag, so
every `v*` tag needs a `## <tag>` heading here before it is pushed.

## v0.0.2

### Fixed

- Empty-text guard wrote a 400 but did not return, so the empty message was
  still queued and delivered.
- Notification intake blocked forever when the send queue was full; it now sheds
  load with a 503 instead of holding the connection open.
- Message text is escaped for Telegram's HTML parse mode. Because delivery is
  asynchronous, a subject containing `<`, `>` or `&` was previously rejected by
  Telegram and dropped, while the caller had already received a 200. Callers
  that want real markup send `"format": "html"`.
- The per-chat password is compared in constant time.
- Startup fails closed: a missing `TELEGRAM_TOKEN`, `TRANSFORMER_SEED`,
  `API_DOMAIN` or `ADMIN_TELEGRAM_ID` now exits non-zero listing everything that
  is missing, instead of continuing with a nil config or panicking later.
- Shutdown no longer races: the HTTP server drains before the notification
  channel is closed, so queued messages are delivered and a send on a closed
  channel cannot panic.
- `Stop()` no longer panics when the bot handler failed to start.
- The unchecked `strconv.ParseInt` error in the sender would have sent to chat
  id 0.
- `/getbashscript` no longer sends an empty message when the template fails.
- Unset `LOG_LEVEL` means info, not debug.

### Added

- `GET /healthz`, plus a `--healthcheck` subcommand for the container probe.
- Request body size limit and a read-header timeout.
- `format` field on the notification payload (`text`, the default, or `html`).

### Changed

- Multi-stage, non-root, distroless image: ~1 GB down to ~22 MB, with a
  `HEALTHCHECK`. The image source label pointed at a repo this was never built
  from.
- CI replaced with the standard lint + test + build workflow, and releases now
  go through the shared build-push-deploy pipeline.
