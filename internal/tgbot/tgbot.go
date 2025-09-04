package tgbot

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type tgBot struct {
	ctx context.Context
	b   *bot.Bot
	log *slog.Logger
}

func NewTgBot(ctx context.Context, token string) (*tgBot, error) {
	var err error

	tgBot := &tgBot{
		ctx: ctx,
		log: slog.Default(),
	}

	tgBot.b, err = bot.New(
		token,
		bot.WithDefaultHandler(tgBot.updateHandler),
	)
	if err != nil {
		return tgBot, err
	}

	return tgBot, nil
}

func (b *tgBot) Start() {
	go b.b.Start(b.ctx)
}

func (b *tgBot) updateHandler(ctx context.Context, _ *bot.Bot, _ *models.Update) {}
