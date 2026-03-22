package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func main() {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	content := NewContentStore(cfg)
	if err := content.ValidateForMode(cfg.Mode); err != nil {
		log.Fatalf("validate content: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	opts := []tgbot.Option{
		tgbot.WithErrorsHandler(func(err error) {
			log.Printf("telegram error: %v", err)
		}),
	}
	if cfg.Debug {
		opts = append(opts, tgbot.WithDebug())
	}

	client, err := tgbot.New(cfg.BotToken, opts...)
	if err != nil {
		log.Fatalf("create bot: %v", err)
	}

	app := NewMaintenanceApp(cfg, client, content, time.Now)
	client.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return true
	}, app.HandleUpdate)

	if _, err := client.DeleteWebhook(ctx, &tgbot.DeleteWebhookParams{DropPendingUpdates: true}); err != nil {
		log.Printf("delete webhook: %v", err)
	}

	log.Printf("starting maintenance bot in %s mode with static dir %s", cfg.Mode, cfg.StaticDir)
	client.Start(ctx)
	log.Printf("maintenance bot stopped")
}
