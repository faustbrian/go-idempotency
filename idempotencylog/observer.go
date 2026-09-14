// Package idempotencylog is the legacy structured-log adapter.
//
// Deprecated: use github.com/faustbrian/go-idempotency/adapters/slog. This
// package remains supported for the longer of 180 days after successor
// availability and two subsequently published stable root-module minor
// releases.
package idempotencylog

import (
	"context"
	"log/slog"

	"github.com/faustbrian/go-idempotency"
	canonical "github.com/faustbrian/go-idempotency/adapters/slog"
)

// ErrNilLogger reports an unusable logger configuration.
var ErrNilLogger = canonical.ErrNilLogger

// Observer preserves the legacy slog adapter type identity.
type Observer struct{ inner *canonical.Observer }

// New constructs an observer for a standard slog logger.
func New(logger *slog.Logger) (*Observer, error) {
	inner, err := canonical.New(logger)
	if err != nil {
		return nil, err
	}

	return &Observer{inner: inner}, nil
}

// Observe writes one bounded transition record.
func (observer *Observer) Observe(ctx context.Context, event idempotency.Observation) {
	observer.inner.Observe(ctx, event)
}

var _ idempotency.Observer = (*Observer)(nil)
