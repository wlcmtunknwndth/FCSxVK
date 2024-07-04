package telegram

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/wlcmtunknwndth/FCSxVK/lib/sl"
	"log/slog"
)

const scope = "internal.telegram."

type AI interface {
	HandlePrompt(ctx context.Context, msg string) (string, error)
}

type Telegram struct {
	bot *bot.Bot
	log *slog.Logger
	ai  AI
}

func New(token string, log *slog.Logger, client AI) (*Telegram, error) {
	const op = scope + "New"

	opts := []bot.Option{
		bot.WithDebug(),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tg := &Telegram{
		bot: b,
		log: log,
		ai:  client,
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/msg", bot.MatchTypePrefix, tg.handler)

	return tg, nil
}

func (t *Telegram) Start(ctx context.Context) {
	t.bot.Start(ctx)
}

func (t *Telegram) Close(ctx context.Context) error {
	const op = scope + "Close"
	if ok, err := t.bot.Close(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	} else if !ok {
		return nil
	}
	return nil
}

func (t *Telegram) handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	const op = scope + "ai_text_handler"

	var resp string
	var err error
	if update.Message != nil && update.Message.Text != "" {
		resp, err = t.ai.HandlePrompt(ctx, update.Message.Text)
		if err != nil {
			t.log.Error("couldn't handle prompt", sl.Op(op), sl.Err(err))
			return
		}
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   models.Message{Text: resp}.Text,
		//ParseMode: models.ParseModeMarkdown,
		//essageParams{ChatID: update.Message.Chat.ID, Text: *resp}.Text,
	})
	if err != nil {
		t.log.Error("couldn't send message", sl.Op(op), sl.Err(err))
		return
	}
	return
}
