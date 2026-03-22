package main

import (
	"testing"
	"time"
)

func TestCooldownStoreAllow(t *testing.T) {
	now := time.Unix(100, 0)
	store := NewCooldownStore(time.Minute, func() time.Time {
		return now
	})

	if !store.Allow(1) {
		t.Fatal("first Allow() = false, want true")
	}
	if store.Allow(1) {
		t.Fatal("second Allow() = true, want false")
	}

	now = now.Add(time.Minute)
	if !store.Allow(1) {
		t.Fatal("Allow() after cooldown = false, want true")
	}
}

func TestCooldownStoreDisabledWhenDurationNonPositive(t *testing.T) {
	store := NewCooldownStore(0, time.Now)

	if !store.Allow(1) || !store.Allow(1) {
		t.Fatal("Allow() = false with disabled cooldown, want always true")
	}
}
