// Package idempotencytest provides reusable adapter conformance and fixtures.
package idempotencytest

import (
	"fmt"
	"sync"
	"time"
)

// Clock is a concurrency-safe deterministic clock for store tests.
type Clock struct {
	mu  sync.Mutex
	now time.Time
}

// NewClock constructs a deterministic clock at now.
func NewClock(now time.Time) *Clock {
	return &Clock{now: now}
}

// Now returns the current deterministic instant.
func (c *Clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// Advance moves the clock forward by duration.
func (c *Clock) Advance(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(duration)
}

// Set replaces the current deterministic instant.
func (c *Clock) Set(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = now
}

// TokenSource emits deterministic unique ownership tokens for tests.
type TokenSource struct {
	mu     sync.Mutex
	prefix string
	next   uint64
}

// NewTokenSource constructs a token source using prefix.
func NewTokenSource(prefix string) *TokenSource {
	return &TokenSource{prefix: prefix}
}

// Next returns the next deterministic token.
func (s *TokenSource) Next() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	return fmt.Sprintf("%s-%d", s.prefix, s.next), nil
}
