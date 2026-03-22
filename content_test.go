package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestContentStoreMessageTextHotReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "message.txt")

	writeFileWithTimestamp(t, path, "first", time.Unix(100, 0))

	store := NewContentStore(Config{
		StaticDir: dir,
		TextFile:  "message.txt",
	})

	content, err := store.MessageText()
	if err != nil {
		t.Fatalf("MessageText() error = %v", err)
	}
	if content != "first" {
		t.Fatalf("MessageText() = %q, want first", content)
	}

	writeFileWithTimestamp(t, path, "second", time.Unix(200, 0))

	content, err = store.MessageText()
	if err != nil {
		t.Fatalf("MessageText() second error = %v", err)
	}
	if content != "second" {
		t.Fatalf("MessageText() = %q, want second", content)
	}
}

func TestContentStoreMediaSpecReloadsFingerprint(t *testing.T) {
	dir := t.TempDir()
	writeFileWithTimestamp(t, filepath.Join(dir, "caption.txt"), "caption one", time.Unix(100, 0))
	writeFileWithTimestamp(t, filepath.Join(dir, "maintenance.jpg"), "a", time.Unix(100, 0))

	store := NewContentStore(Config{
		StaticDir:   dir,
		CaptionFile: "caption.txt",
		MediaFile:   "maintenance.jpg",
	})

	first, err := store.MediaSpec()
	if err != nil {
		t.Fatalf("MediaSpec() error = %v", err)
	}

	writeFileWithTimestamp(t, filepath.Join(dir, "caption.txt"), "caption two", time.Unix(200, 0))
	writeFileWithTimestamp(t, filepath.Join(dir, "maintenance.jpg"), "ab", time.Unix(300, 0))

	second, err := store.MediaSpec()
	if err != nil {
		t.Fatalf("MediaSpec() second error = %v", err)
	}

	if first.Fingerprint == second.Fingerprint {
		t.Fatalf("Fingerprint did not change: %q", first.Fingerprint)
	}
	if second.Caption != "caption two" {
		t.Fatalf("Caption = %q, want caption two", second.Caption)
	}
}

func TestContentStoreValidateForModeFailsWhenFilesMissing(t *testing.T) {
	dir := t.TempDir()
	store := NewContentStore(Config{
		StaticDir: dir,
		TextFile:  "message.txt",
	})

	err := store.ValidateForMode(ModeText)
	if err == nil {
		t.Fatal("ValidateForMode() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "message.txt") {
		t.Fatalf("error = %v, want file path mention", err)
	}
}

func writeFileWithTimestamp(t *testing.T, path, body string, ts time.Time) {
	t.Helper()

	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
	if err := os.Chtimes(path, ts, ts); err != nil {
		t.Fatalf("Chtimes(%s): %v", path, err)
	}
}
