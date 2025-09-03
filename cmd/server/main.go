package main

import (
	"15-puzzle/internal/repo"
	"15-puzzle/internal/tgbot"
	"15-puzzle/internal/web-service/handler"
	"15-puzzle/internal/web-service/server"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	token := requireEnv("BOT_TOKEN")

	bot, err := tgbot.NewTgBot(ctx, token)
	if err != nil {
		exitWithError("bot init: %s", err)
	}
	bot.Start()

	r, err := repo.NewFileRepo(ctx, requireEnv("DATA_FILE"))
	if err != nil {
		exitWithError("repo init: %s", err)
	}

	server.StartServer(ctx,
		handler.NewHandler(r, token, requireEnv("ACCESS_CODE"), os.Getenv("CONTEXT_ROOT"), os.Getenv("STATIC_DIR"), requireEnv("PROJECT_LINK")))
}

func requireEnv(env string) string {
	v, ok := os.LookupEnv(env)
	if !ok {
		exitWithError(env + " variable not set")
	}
	return v
}

func exitWithError(format string, a ...any) {
	slog.Error(fmt.Sprintf(format, a...))
	os.Exit(1)
}
