package main

import (
	"strings"
	"testing"
	"time"
)

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("BOT_TOKEN", "token")
	t.Setenv("PLACEHOLDER_MODE", "")
	t.Setenv("STATIC_DIR", "")
	t.Setenv("TEXT_FILE", "")
	t.Setenv("CAPTION_FILE", "")
	t.Setenv("MEDIA_FILE", "")
	t.Setenv("REPLY_COOLDOWN", "")
	t.Setenv("DEBUG", "")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv() error = %v", err)
	}

	if cfg.Mode != ModeText {
		t.Fatalf("Mode = %q, want %q", cfg.Mode, ModeText)
	}
	if cfg.StaticDir != "/opt/app/static" {
		t.Fatalf("StaticDir = %q", cfg.StaticDir)
	}
	if cfg.TextFile != "message.txt" {
		t.Fatalf("TextFile = %q", cfg.TextFile)
	}
	if cfg.CaptionFile != "caption.txt" {
		t.Fatalf("CaptionFile = %q", cfg.CaptionFile)
	}
	if cfg.ReplyCooldown != time.Minute {
		t.Fatalf("ReplyCooldown = %s", cfg.ReplyCooldown)
	}
	if cfg.Debug {
		t.Fatalf("Debug = true, want false")
	}
}

func TestLoadConfigMediaRequiresFile(t *testing.T) {
	t.Setenv("BOT_TOKEN", "token")
	t.Setenv("PLACEHOLDER_MODE", "photo")
	t.Setenv("MEDIA_FILE", "")

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatal("LoadConfigFromEnv() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "MEDIA_FILE") {
		t.Fatalf("error = %v, want MEDIA_FILE mention", err)
	}
}

func TestLoadConfigRejectsEscapingStaticDir(t *testing.T) {
	t.Setenv("BOT_TOKEN", "token")
	t.Setenv("TEXT_FILE", "../message.txt")

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatal("LoadConfigFromEnv() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "TEXT_FILE") {
		t.Fatalf("error = %v, want TEXT_FILE mention", err)
	}
}
