package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync/atomic"
	"time"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type MaintenanceApp struct {
	cfg      Config
	client   TelegramClient
	sender   *PlaceholderSender
	cooldown *CooldownStore
	started  time.Time
	now      func() time.Time

	updates uint64
}

func NewMaintenanceApp(cfg Config, client TelegramClient, content *ContentStore, now func() time.Time) *MaintenanceApp {
	if now == nil {
		now = time.Now
	}

	return &MaintenanceApp{
		cfg:      cfg,
		client:   client,
		sender:   NewPlaceholderSender(client, content, cfg.Mode),
		cooldown: NewCooldownStore(cfg.ReplyCooldown, now),
		started:  now(),
		now:      now,
	}
}

func (a *MaintenanceApp) HandleUpdate(ctx context.Context, _ *tgbot.Bot, update *models.Update) {
	if update == nil || update.Message == nil {
		return
	}

	atomic.AddUint64(&a.updates, 1)

	chatID := update.Message.Chat.ID
	if isStatusCommand(update.Message.Text) {
		if err := a.sendStatus(ctx, chatID); err != nil {
			log.Printf("send status: %v", err)
		}
		return
	}

	if !a.cooldown.Allow(chatID) {
		return
	}

	if err := a.sender.Send(ctx, chatID); err != nil {
		log.Printf("send placeholder to chat %d: %v", chatID, err)
	}
}

func (a *MaintenanceApp) sendStatus(ctx context.Context, chatID int64) error {
	_, err := a.client.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:    chatID,
		Text:      a.statusText(),
		ParseMode: models.ParseModeHTML,
	})
	return err
}

func (a *MaintenanceApp) statusText() string {
	uptime := a.now().Sub(a.started).Round(time.Second)
	if uptime < 0 {
		uptime = 0
	}

	return fmt.Sprintf(
		"Uptime: %s\nUpdates: %d\nMode: %s\nCooldown: %s\nMedia cache: %s",
		uptime,
		atomic.LoadUint64(&a.updates),
		a.cfg.Mode,
		a.cfg.ReplyCooldown,
		a.sender.MediaCacheStatus(),
	)
}

func isStatusCommand(text string) bool {
	firstField := strings.TrimSpace(text)
	if firstField == "" {
		return false
	}

	command := strings.Fields(firstField)[0]
	command = strings.SplitN(command, "@", 2)[0]

	return command == "/c" || command == "/status"
}
