package main

import (
	"context"
	"flag"
	"github.com/wlcmtunknwndth/FCSxVK/internal/AI/gemini"
	"github.com/wlcmtunknwndth/FCSxVK/internal/telegram"
	"github.com/wlcmtunknwndth/FCSxVK/lib/sl"
	"log/slog"
	"os"
	"os/signal"
)

const (
	botApiEnv = "api_token"

	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"

	scope = "main"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var botToken string
	flag.StringVar(&botToken, "token", "", "bot token")

	var programEnv string
	flag.StringVar(&programEnv, "env", "local", "run environment")

	var aiToken string
	flag.StringVar(&aiToken, "ai_token", "", "AI api token")

	var proxyUrl string
	flag.StringVar(&proxyUrl, "proxy", "", "proxy server url")

	var username string
	flag.StringVar(&username, "user", "", "proxy username")

	var password string
	flag.StringVar(&password, "pass", "", "proxy password")

	var staticPath string
	flag.StringVar(&staticPath, "static", "", "route to file storage")
	flag.Parse()

	log := setupLogger(programEnv)

	ai, err := gemini.New(ctx, aiToken, proxyUrl, username, password)
	if err != nil {
		log.Error("couldn't connect to gemini api", sl.Op(scope), sl.Err(err))
		return
	}

	bot, err := telegram.New(botToken, log, ai, staticPath)
	if err != nil {
		log.Error("couldn't create telegram bot instance", sl.Op(scope), sl.Err(err))
		return
	}
	log.Info("initialized bot")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	go bot.Start(ctx)

	<-stop

	err = bot.Close(ctx)
	if err != nil {
		log.Error("couldn't close bot", sl.Op(scope), sl.Err(err))
		return
	}
	//resp, err := ai.HandleTextPrompt(ctx, "hello, gemini. What can you do?")
	//if err != nil {
	//	slog.Error("couldn't handle prompt", sl.Op(scope), sl.Err(err))
	//	return
	//}

	//log.Info("got:", slog.Any("resp", resp), slog.Int("len", len(resp)))

	return
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
