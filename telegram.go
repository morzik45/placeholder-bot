package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type TelegramClient interface {
	SendMessage(ctx context.Context, params *tgbot.SendMessageParams) (*models.Message, error)
	SendPhoto(ctx context.Context, params *tgbot.SendPhotoParams) (*models.Message, error)
	SendVideo(ctx context.Context, params *tgbot.SendVideoParams) (*models.Message, error)
}

type mediaCache struct {
	fingerprint string
	fileID      string
}

type PlaceholderSender struct {
	client  TelegramClient
	content *ContentStore
	mode    PlaceholderMode

	mu    sync.Mutex
	cache mediaCache
}

func NewPlaceholderSender(client TelegramClient, content *ContentStore, mode PlaceholderMode) *PlaceholderSender {
	return &PlaceholderSender{
		client:  client,
		content: content,
		mode:    mode,
	}
}

func (s *PlaceholderSender) Send(ctx context.Context, chatID int64) error {
	switch s.mode {
	case ModeText:
		text, err := s.content.MessageText()
		if err != nil {
			return err
		}
		_, err = s.client.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseModeHTML,
		})
		return err
	case ModePhoto, ModeVideo:
		return s.sendMedia(ctx, chatID, s.mode)
	default:
		return fmt.Errorf("unsupported mode %q", s.mode)
	}
}

func (s *PlaceholderSender) MediaCacheStatus() string {
	if s.mode == ModeText {
		return "disabled"
	}

	spec, err := s.content.MediaSpec()
	if err != nil {
		return fmt.Sprintf("error (%v)", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cache.fingerprint == spec.Fingerprint && s.cache.fileID != "" {
		return fmt.Sprintf("warm (%s)", shortFingerprint(spec.Fingerprint))
	}
	return fmt.Sprintf("cold (%s)", shortFingerprint(spec.Fingerprint))
}

func (s *PlaceholderSender) sendMedia(ctx context.Context, chatID int64, mode PlaceholderMode) error {
	spec, err := s.content.MediaSpec()
	if err != nil {
		return err
	}

	if fileID, ok := s.cachedFileID(spec.Fingerprint); ok {
		if err := s.sendMediaByFileID(ctx, chatID, mode, spec.Caption, fileID); err == nil {
			return nil
		}
		s.resetCache(spec.Fingerprint)
	}

	return s.sendMediaByUpload(ctx, chatID, mode, spec)
}

func (s *PlaceholderSender) sendMediaByFileID(ctx context.Context, chatID int64, mode PlaceholderMode, caption, fileID string) error {
	input := &models.InputFileString{Data: fileID}

	switch mode {
	case ModePhoto:
		_, err := s.client.SendPhoto(ctx, &tgbot.SendPhotoParams{
			ChatID:    chatID,
			Photo:     input,
			Caption:   caption,
			ParseMode: models.ParseModeHTML,
		})
		return err
	case ModeVideo:
		_, err := s.client.SendVideo(ctx, &tgbot.SendVideoParams{
			ChatID:            chatID,
			Video:             input,
			Caption:           caption,
			ParseMode:         models.ParseModeHTML,
			SupportsStreaming: true,
		})
		return err
	default:
		return fmt.Errorf("unsupported media mode %q", mode)
	}
}

func (s *PlaceholderSender) sendMediaByUpload(ctx context.Context, chatID int64, mode PlaceholderMode, spec MediaSpec) error {
	file, err := os.Open(spec.Path)
	if err != nil {
		return fmt.Errorf("open media file %s: %w", spec.Path, err)
	}
	defer file.Close()

	input := &models.InputFileUpload{
		Filename: spec.Filename,
		Data:     file,
	}

	var message *models.Message
	switch mode {
	case ModePhoto:
		message, err = s.client.SendPhoto(ctx, &tgbot.SendPhotoParams{
			ChatID:    chatID,
			Photo:     input,
			Caption:   spec.Caption,
			ParseMode: models.ParseModeHTML,
		})
	case ModeVideo:
		message, err = s.client.SendVideo(ctx, &tgbot.SendVideoParams{
			ChatID:            chatID,
			Video:             input,
			Caption:           spec.Caption,
			ParseMode:         models.ParseModeHTML,
			SupportsStreaming: true,
		})
	default:
		return fmt.Errorf("unsupported media mode %q", mode)
	}
	if err != nil {
		return err
	}

	if fileID := extractMediaFileID(mode, message); fileID != "" {
		s.storeCache(spec.Fingerprint, fileID)
	}

	return nil
}

func (s *PlaceholderSender) cachedFileID(fingerprint string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cache.fingerprint != fingerprint || s.cache.fileID == "" {
		return "", false
	}

	return s.cache.fileID, true
}

func (s *PlaceholderSender) storeCache(fingerprint, fileID string) {
	s.mu.Lock()
	s.cache = mediaCache{
		fingerprint: fingerprint,
		fileID:      fileID,
	}
	s.mu.Unlock()
}

func (s *PlaceholderSender) resetCache(fingerprint string) {
	s.mu.Lock()
	if s.cache.fingerprint == fingerprint {
		s.cache = mediaCache{}
	}
	s.mu.Unlock()
}

func extractMediaFileID(mode PlaceholderMode, message *models.Message) string {
	if message == nil {
		return ""
	}

	switch mode {
	case ModePhoto:
		if len(message.Photo) == 0 {
			return ""
		}
		return message.Photo[len(message.Photo)-1].FileID
	case ModeVideo:
		if message.Video == nil {
			return ""
		}
		return message.Video.FileID
	default:
		return ""
	}
}
