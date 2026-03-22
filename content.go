package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type MediaSpec struct {
	Path        string
	Filename    string
	Caption     string
	Fingerprint string
}

type fileFingerprint struct {
	path    string
	size    int64
	modTime time.Time
}

type cachedTextFile struct {
	ready       bool
	fingerprint fileFingerprint
	content     string
}

type ContentStore struct {
	cfg Config

	mu      sync.Mutex
	text    cachedTextFile
	caption cachedTextFile
}

func NewContentStore(cfg Config) *ContentStore {
	return &ContentStore{cfg: cfg}
}

func (s *ContentStore) ValidateForMode(mode PlaceholderMode) error {
	switch mode {
	case ModeText:
		_, err := s.MessageText()
		return err
	case ModePhoto, ModeVideo:
		_, err := s.MediaSpec()
		return err
	default:
		return fmt.Errorf("unsupported mode %q", mode)
	}
}

func (s *ContentStore) MessageText() (string, error) {
	return s.loadTextFile(s.cfg.TextPath(), &s.text)
}

func (s *ContentStore) Caption() (string, error) {
	return s.loadTextFile(s.cfg.CaptionPath(), &s.caption)
}

func (s *ContentStore) MediaSpec() (MediaSpec, error) {
	caption, err := s.Caption()
	if err != nil {
		return MediaSpec{}, err
	}

	path := s.cfg.MediaPath()
	info, err := os.Stat(path)
	if err != nil {
		return MediaSpec{}, fmt.Errorf("stat media file %s: %w", path, err)
	}
	if info.IsDir() {
		return MediaSpec{}, fmt.Errorf("media path %s is a directory", path)
	}

	return MediaSpec{
		Path:        path,
		Filename:    filepath.Base(path),
		Caption:     caption,
		Fingerprint: makeFingerprint(path, info).String(),
	}, nil
}

func (s *ContentStore) loadTextFile(path string, cache *cachedTextFile) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("stat text file %s: %w", path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("text path %s is a directory", path)
	}

	fingerprint := makeFingerprint(path, info)

	s.mu.Lock()
	if cache.ready && cache.fingerprint == fingerprint {
		content := cache.content
		s.mu.Unlock()
		return content, nil
	}
	s.mu.Unlock()

	body, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read text file %s: %w", path, err)
	}

	s.mu.Lock()
	cache.ready = true
	cache.fingerprint = fingerprint
	cache.content = string(body)
	content := cache.content
	s.mu.Unlock()

	return content, nil
}

func makeFingerprint(path string, info os.FileInfo) fileFingerprint {
	return fileFingerprint{
		path:    path,
		size:    info.Size(),
		modTime: info.ModTime().UTC(),
	}
}

func (f fileFingerprint) String() string {
	return fmt.Sprintf("%s|%d|%d", filepath.Base(f.path), f.size, f.modTime.UnixNano())
}

func shortFingerprint(value string) string {
	if value == "" {
		return "empty"
	}
	if len(value) <= 32 {
		return value
	}
	return value[:32]
}

func (s *ContentStore) EnsureStaticDir() error {
	info, err := os.Stat(s.cfg.StaticDir)
	if err != nil {
		return fmt.Errorf("stat STATIC_DIR %s: %w", s.cfg.StaticDir, err)
	}
	if !info.IsDir() {
		return errors.New("STATIC_DIR must point to a directory")
	}
	return nil
}
