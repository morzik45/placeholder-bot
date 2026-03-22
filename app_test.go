package main

import (
	"context"
	"testing"
	"time"
)

func TestMaintenanceAppStatusUsesHTMLParseMode(t *testing.T) {
	client := &fakeTelegramClient{}
	app := NewMaintenanceApp(Config{
		Mode:          ModeText,
		ReplyCooldown: time.Minute,
	}, client, NewContentStore(Config{}), func() time.Time {
		return time.Unix(100, 0)
	})

	if err := app.sendStatus(context.Background(), 1); err != nil {
		t.Fatalf("sendStatus() error = %v", err)
	}

	if client.lastMessage == nil {
		t.Fatal("lastMessage = nil, want sent status")
	}
	if client.lastMessage.ParseMode != "HTML" {
		t.Fatalf("ParseMode = %q, want HTML", client.lastMessage.ParseMode)
	}
}
