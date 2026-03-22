package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type PlaceholderMode string

const (
	ModeText  PlaceholderMode = "text"
	ModePhoto PlaceholderMode = "photo"
	ModeVideo PlaceholderMode = "video"
)

type Config struct {
	BotToken      string
	Mode          PlaceholderMode
	StaticDir     string
	TextFile      string
	CaptionFile   string
	MediaFile     string
	ReplyCooldown time.Duration
	Debug         bool
}

func LoadConfigFromEnv() (Config, error) {
	cfg := Config{
		BotToken:      strings.TrimSpace(os.Getenv("BOT_TOKEN")),
		Mode:          PlaceholderMode(strings.ToLower(strings.TrimSpace(getEnv("PLACEHOLDER_MODE", string(ModeText))))),
		StaticDir:     strings.TrimSpace(getEnv("STATIC_DIR", "/opt/app/static")),
		TextFile:      strings.TrimSpace(getEnv("TEXT_FILE", "message.txt")),
		CaptionFile:   strings.TrimSpace(getEnv("CAPTION_FILE", "caption.txt")),
		MediaFile:     strings.TrimSpace(os.Getenv("MEDIA_FILE")),
		ReplyCooldown: time.Minute,
	}

	if rawCooldown := strings.TrimSpace(os.Getenv("REPLY_COOLDOWN")); rawCooldown != "" {
		cooldown, err := time.ParseDuration(rawCooldown)
		if err != nil {
			return Config{}, fmt.Errorf("parse REPLY_COOLDOWN: %w", err)
		}
		cfg.ReplyCooldown = cooldown
	}

	if rawDebug := strings.TrimSpace(os.Getenv("DEBUG")); rawDebug != "" {
		debug, err := strconv.ParseBool(rawDebug)
		if err != nil {
			return Config{}, fmt.Errorf("parse DEBUG: %w", err)
		}
		cfg.Debug = debug
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	textFile, _ := normalizeStaticPath(cfg.TextFile)
	captionFile, _ := normalizeStaticPath(cfg.CaptionFile)
	mediaFile, _ := normalizeStaticPath(cfg.MediaFile)
	cfg.TextFile = textFile
	cfg.CaptionFile = captionFile
	cfg.MediaFile = mediaFile

	return cfg, nil
}

func (c Config) Validate() error {
	if c.BotToken == "" {
		return errors.New("BOT_TOKEN is required")
	}

	if c.StaticDir == "" {
		return errors.New("STATIC_DIR must not be empty")
	}

	if !c.Mode.Valid() {
		return fmt.Errorf("unsupported PLACEHOLDER_MODE %q", c.Mode)
	}

	textFile, err := normalizeStaticPath(c.TextFile)
	if err != nil {
		return fmt.Errorf("TEXT_FILE: %w", err)
	}
	captionFile, err := normalizeStaticPath(c.CaptionFile)
	if err != nil {
		return fmt.Errorf("CAPTION_FILE: %w", err)
	}
	mediaFile := c.MediaFile
	if mediaFile != "" {
		mediaFile, err = normalizeStaticPath(mediaFile)
		if err != nil {
			return fmt.Errorf("MEDIA_FILE: %w", err)
		}
	}

	if c.Mode == ModeText {
		if textFile == "" {
			return errors.New("TEXT_FILE must not be empty")
		}
	} else {
		if captionFile == "" {
			return errors.New("CAPTION_FILE must not be empty")
		}
		if mediaFile == "" {
			return errors.New("MEDIA_FILE is required for photo and video modes")
		}
	}

	return nil
}

func (c Config) TextPath() string {
	return filepath.Join(c.StaticDir, c.TextFile)
}

func (c Config) CaptionPath() string {
	return filepath.Join(c.StaticDir, c.CaptionFile)
}

func (c Config) MediaPath() string {
	return filepath.Join(c.StaticDir, c.MediaFile)
}

func (m PlaceholderMode) Valid() bool {
	switch m {
	case ModeText, ModePhoto, ModeVideo:
		return true
	default:
		return false
	}
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func normalizeStaticPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", nil
	}

	cleaned := filepath.Clean(path)
	if filepath.IsAbs(cleaned) {
		return "", errors.New("must be relative to STATIC_DIR")
	}
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", errors.New("must stay inside STATIC_DIR")
	}

	return cleaned, nil
}
