package telegram

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/wlcmtunknwndth/FCSxVK/lib/get"
	"github.com/wlcmtunknwndth/FCSxVK/lib/sl"
	"log/slog"
	"strings"
)

const (
	scope = "internal.telegram."

	textOnlyResp     = "/text"
	imageAndTextResp = "/image"

	promtMardownSyntax = "Send me text without markdown syntax, please"
)

type AI interface {
	HandleTextPrompt(ctx context.Context, msg string) (string, error)
	HandleTextAndImagePrompt(ctx context.Context, filePath, msgPrompt string) (string, error)
}

type Telegram struct {
	bot        *bot.Bot
	log        *slog.Logger
	ai         AI
	staticPath string
}

func New(token string, log *slog.Logger, client AI, static string) (*Telegram, error) {
	const op = scope + "New"

	opts := []bot.Option{
		bot.WithDebug(),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tg := &Telegram{
		bot:        b,
		log:        log,
		ai:         client,
		staticPath: static,
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, textOnlyResp, bot.MatchTypePrefix, tg.textHandler)
	//b.RegisterHandler(bot.HandlerTypeMessageText, imageAndTextResp, bot.MatchTypePrefix, tg.imageAndTextHandler)
	b.RegisterHandlerMatchFunc(matchCaption, tg.imageAndTextHandler)
	return tg, nil
}

func matchCaption(update *models.Update) bool {
	if update.Message.Caption[:len(imageAndTextResp)] == imageAndTextResp {
		return true
	}
	return false
}

//matchFunc := func(update *models.Update) bool {
//	// your checks
//	return true
//}

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

func (t *Telegram) imageAndTextHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	const op = scope + "imageAndTextHandler"

	var fileInfo *models.File
	var err error

	if update.Message.Photo != nil {
		fileInfo, err = b.GetFile(ctx, &bot.GetFileParams{FileID: update.Message.Photo[0].FileID})
		if err != nil {
			t.log.Error("couldn't get image", sl.Op(op), sl.Err(err))
			return
		}
	} else {
		if err = t.sendMessage(ctx, b, update, "no documents sent"); err != nil {
			t.log.Error("couldn't send message", sl.Op(op), sl.Err(err))
			return
		}
		return
	}

	link := b.FileDownloadLink(fileInfo)
	//link := b.FileDownloadLink(update.Message.Document)
	split := strings.Split(fileInfo.FilePath, ".")
	if len(split) < 2 {
		t.log.Error("couldn't get file extension", sl.Op(op), sl.Err(err))
		return
	}
	path := fmt.Sprintf("%s/%s.%s", t.staticPath, update.Message.Photo[0].FileUniqueID, split[len(split)-1])

	err = get.DownloadFile(path, link)
	if err != nil {
		t.log.Error("couldn't download file", sl.Op(op), sl.Err(err))
		return
	}

	msg, err := t.ai.HandleTextAndImagePrompt(ctx, path, update.Message.Caption[len(imageAndTextResp)-1:])
	if err != nil {
		t.log.Error("couldn't get response", sl.Op(op), sl.Err(err))
		return
	}

	err = t.sendMessage(ctx, b, update, msg)
	if err != nil {
		t.log.Error("couldn't send message", sl.Op(op), sl.Err(err))
		return
	}
}

func (t *Telegram) textHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	const op = scope + "textHandler"

	var resp string
	var err error
	cleanPrompt := update.Message.Text[len(textOnlyResp)-1:]
	if update.Message != nil && update.Message.Text != "" {
		resp, err = t.ai.HandleTextPrompt(ctx, cleanPrompt)
		if err != nil {
			t.log.Error("couldn't handle prompt", sl.Op(op), sl.Err(err))
			return
		}
	}

	//_, err = b.SendMessage(ctx, &bot.SendMessageParams{
	//	ChatID: update.Message.Chat.ID,
	//	Text:   models.Message{Text: resp}.Text,
	//	//ParseMode: models.ParseModeMarkdown,
	//})
	//if err != nil {
	//	t.log.Error("couldn't send message", sl.Op(op), sl.Err(err))
	//	return
	//}
	err = t.sendMessage(ctx, b, update, resp)
	if err != nil {
		return
	}
	return
}

func (t *Telegram) sendMessage(ctx context.Context, b *bot.Bot, update *models.Update, msg string) error {
	const op = scope + "sendMessage"

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   models.Message{Text: msg}.Text,
		//ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
