FROM golang:latest

LABEL org.opencontainers.image.source=https://github.com/supercakecrumb/curly-bot
LABEL org.opencontainers.image.description="Curly telegram bot"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . ./

RUN go build -o main ./cmd

CMD ["/app/main"]