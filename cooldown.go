package main

import (
	"sync"
	"time"
)

type CooldownStore struct {
	duration time.Duration
	now      func() time.Time

	mu       sync.Mutex
	nextSend map[int64]time.Time
}

func NewCooldownStore(duration time.Duration, now func() time.Time) *CooldownStore {
	if now == nil {
		now = time.Now
	}

	return &CooldownStore{
		duration: duration,
		now:      now,
		nextSend: make(map[int64]time.Time),
	}
}

func (s *CooldownStore) Allow(chatID int64) bool {
	if s.duration <= 0 {
		return true
	}

	now := s.now()

	s.mu.Lock()
	defer s.mu.Unlock()

	if next, ok := s.nextSend[chatID]; ok && now.Before(next) {
		return false
	}

	s.nextSend[chatID] = now.Add(s.duration)
	return true
}
