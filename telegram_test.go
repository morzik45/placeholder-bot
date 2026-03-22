package main

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func TestPlaceholderSenderCachesUploadedPhotoFileID(t *testing.T) {
	dir := t.TempDir()
	writeFileWithTimestamp(t, filepath.Join(dir, "caption.txt"), "caption", time.Unix(100, 0))
	writeFileWithTimestamp(t, filepath.Join(dir, "image.jpg"), "img", time.Unix(100, 0))

	client := &fakeTelegramClient{}
	sender := NewPlaceholderSender(client, NewContentStore(Config{
		StaticDir:   dir,
		CaptionFile: "caption.txt",
		MediaFile:   "image.jpg",
	}), ModePhoto)

	if err := sender.Send(context.Background(), 42); err != nil {
		t.Fatalf("Send() first error = %v", err)
	}
	if client.photoUploads != 1 {
		t.Fatalf("photoUploads = %d, want 1", client.photoUploads)
	}
	if client.photoByID != 0 {
		t.Fatalf("photoByID = %d, want 0", client.photoByID)
	}

	if err := sender.Send(context.Background(), 42); err != nil {
		t.Fatalf("Send() second error = %v", err)
	}
	if client.photoUploads != 1 {
		t.Fatalf("photoUploads = %d, want 1 after cache hit", client.photoUploads)
	}
	if client.photoByID != 1 {
		t.Fatalf("photoByID = %d, want 1", client.photoByID)
	}
}

func TestPlaceholderSenderResetsCacheWhenMediaFingerprintChanges(t *testing.T) {
	dir := t.TempDir()
	imagePath := filepath.Join(dir, "video.mp4")
	writeFileWithTimestamp(t, filepath.Join(dir, "caption.txt"), "caption", time.Unix(100, 0))
	writeFileWithTimestamp(t, imagePath, "v1", time.Unix(100, 0))

	client := &fakeTelegramClient{}
	sender := NewPlaceholderSender(client, NewContentStore(Config{
		StaticDir:   dir,
		CaptionFile: "caption.txt",
		MediaFile:   "video.mp4",
	}), ModeVideo)

	if err := sender.Send(context.Background(), 7); err != nil {
		t.Fatalf("Send() first error = %v", err)
	}

	writeFileWithTimestamp(t, imagePath, "v2", time.Unix(200, 0))

	if err := sender.Send(context.Background(), 7); err != nil {
		t.Fatalf("Send() second error = %v", err)
	}

	if client.videoUploads != 2 {
		t.Fatalf("videoUploads = %d, want 2", client.videoUploads)
	}
	if client.videoByID != 0 {
		t.Fatalf("videoByID = %d, want 0 because fingerprint changed", client.videoByID)
	}
}

func TestPlaceholderSenderFallsBackToUploadWhenCachedFileIDFails(t *testing.T) {
	dir := t.TempDir()
	writeFileWithTimestamp(t, filepath.Join(dir, "caption.txt"), "caption", time.Unix(100, 0))
	writeFileWithTimestamp(t, filepath.Join(dir, "image.jpg"), "img", time.Unix(100, 0))

	client := &fakeTelegramClient{
		failNextPhotoByID: true,
	}
	sender := NewPlaceholderSender(client, NewContentStore(Config{
		StaticDir:   dir,
		CaptionFile: "caption.txt",
		MediaFile:   "image.jpg",
	}), ModePhoto)

	if err := sender.Send(context.Background(), 42); err != nil {
		t.Fatalf("Send() first error = %v", err)
	}
	if err := sender.Send(context.Background(), 42); err != nil {
		t.Fatalf("Send() second error = %v", err)
	}

	if client.photoByID != 1 {
		t.Fatalf("photoByID = %d, want 1", client.photoByID)
	}
	if client.photoUploads != 2 {
		t.Fatalf("photoUploads = %d, want 2 after retry upload", client.photoUploads)
	}
}

type fakeTelegramClient struct {
	messageCalls int
	photoUploads int
	photoByID    int
	videoUploads int
	videoByID    int

	failNextPhotoByID bool

	lastMessage *tgbot.SendMessageParams
	lastPhoto   *tgbot.SendPhotoParams
	lastVideo   *tgbot.SendVideoParams
}

func (c *fakeTelegramClient) SendMessage(_ context.Context, params *tgbot.SendMessageParams) (*models.Message, error) {
	c.messageCalls++
	c.lastMessage = params
	return &models.Message{}, nil
}

func (c *fakeTelegramClient) SendPhoto(_ context.Context, params *tgbot.SendPhotoParams) (*models.Message, error) {
	c.lastPhoto = params
	switch file := params.Photo.(type) {
	case *models.InputFileString:
		c.photoByID++
		if c.failNextPhotoByID {
			c.failNextPhotoByID = false
			return nil, errors.New("invalid cached file id")
		}
		return &models.Message{
			Photo: []models.PhotoSize{{FileID: file.Data}},
		}, nil
	case *models.InputFileUpload:
		c.photoUploads++
		return &models.Message{
			Photo: []models.PhotoSize{{FileID: "cached-photo-id"}},
		}, nil
	default:
		return nil, errors.New("unexpected photo input type")
	}
}

func (c *fakeTelegramClient) SendVideo(_ context.Context, params *tgbot.SendVideoParams) (*models.Message, error) {
	c.lastVideo = params
	switch file := params.Video.(type) {
	case *models.InputFileString:
		c.videoByID++
		return &models.Message{
			Video: &models.Video{FileID: file.Data},
		}, nil
	case *models.InputFileUpload:
		c.videoUploads++
		return &models.Message{
			Video: &models.Video{FileID: "cached-video-id"},
		}, nil
	default:
		return nil, errors.New("unexpected video input type")
	}
}

func TestPlaceholderSenderUsesHTMLParseModeForText(t *testing.T) {
	dir := t.TempDir()
	writeFileWithTimestamp(t, filepath.Join(dir, "message.txt"), "<b>hello</b>", time.Unix(100, 0))

	client := &fakeTelegramClient{}
	sender := NewPlaceholderSender(client, NewContentStore(Config{
		StaticDir: dir,
		TextFile:  "message.txt",
	}), ModeText)

	if err := sender.Send(context.Background(), 99); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if client.lastMessage == nil {
		t.Fatal("lastMessage = nil, want sent message")
	}
	if client.lastMessage.ParseMode != models.ParseModeHTML {
		t.Fatalf("ParseMode = %q, want %q", client.lastMessage.ParseMode, models.ParseModeHTML)
	}
}

func TestPlaceholderSenderUsesHTMLParseModeForCaptions(t *testing.T) {
	dir := t.TempDir()
	writeFileWithTimestamp(t, filepath.Join(dir, "caption.txt"), "<b>caption</b>", time.Unix(100, 0))
	writeFileWithTimestamp(t, filepath.Join(dir, "image.jpg"), "img", time.Unix(100, 0))

	client := &fakeTelegramClient{}
	sender := NewPlaceholderSender(client, NewContentStore(Config{
		StaticDir:   dir,
		CaptionFile: "caption.txt",
		MediaFile:   "image.jpg",
	}), ModePhoto)

	if err := sender.Send(context.Background(), 42); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if client.lastPhoto == nil {
		t.Fatal("lastPhoto = nil, want sent photo")
	}
	if client.lastPhoto.ParseMode != models.ParseModeHTML {
		t.Fatalf("ParseMode = %q, want %q", client.lastPhoto.ParseMode, models.ParseModeHTML)
	}
}
