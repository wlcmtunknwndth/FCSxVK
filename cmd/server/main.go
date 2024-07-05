package main

import (
	"context"
	"github.com/wlcmtunknwndth/FCSxVK/internal/AI/gemini"
	"github.com/wlcmtunknwndth/FCSxVK/internal/config"
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

	//var botToken string
	//flag.StringVar(&botToken, "token", "", "bot token")
	//
	//var programEnv string
	//flag.StringVar(&programEnv, "env", "local", "run environment")
	//
	//var aiToken string
	//flag.StringVar(&aiToken, "ai_token", "", "AI api token")
	//
	//var proxyUrl string
	//flag.StringVar(&proxyUrl, "proxy", "", "proxy server url")
	//
	//var username string
	//flag.StringVar(&username, "user", "", "proxy username")
	//
	//var password string
	//flag.StringVar(&password, "pass", "", "proxy password")
	//
	//var staticPath string
	//flag.StringVar(&staticPath, "static", "", "route to file storage")
	//flag.Parse()

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	ai, err := gemini.New(ctx, cfg.AiToken, cfg.Proxy.Addr, cfg.Proxy.Username, cfg.Proxy.Password)
	if err != nil {
		log.Error("couldn't connect to gemini api", sl.Op(scope), sl.Err(err))
		return
	}

	bot, err := telegram.New(cfg.TgToken, log, ai, cfg.StaticPath)
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
