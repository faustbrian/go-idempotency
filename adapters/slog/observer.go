// Package idempotencylog adapts idempotency observations to log/slog.
// Loggers constructed by github.com/faustbrian/go-log use this standard type.
package idempotencylog

import (
	"context"
	"errors"
	"log/slog"

	"github.com/faustbrian/go-idempotency/v3"
)

// ErrNilLogger reports an unusable logger configuration.
var ErrNilLogger = errors.New("idempotencylog: nil logger")

// Observer writes semantic transition fields to a slog logger.
type Observer struct {
	logger *slog.Logger
}

// New constructs an observer for a standard slog logger, including loggers
// returned by github.com/faustbrian/go-log.
func New(logger *slog.Logger) (*Observer, error) {
	if logger == nil {
		return nil, ErrNilLogger
	}

	return &Observer{logger: logger}, nil
}

// Observe logs caller-provided observations synchronously, including Correlation
// verbatim. Custom hashers/manual events can expose unbounded sensitive text.
// The application owns handler redaction, blocking and concurrent behavior.
func (observer *Observer) Observe(ctx context.Context, event idempotency.Observation) {
	observer.logger.InfoContext(ctx, "idempotency transition",
		slog.String("transition", string(event.Transition)),
		slog.String("outcome", string(event.Outcome)),
		slog.String("reason", string(event.Reason)),
		slog.Bool("durable", event.Durable),
		slog.String("correlation", event.Correlation),
	)
}

var _ idempotency.Observer = (*Observer)(nil)
