// Package idempotencyqueue is the legacy durable queue adapter.
//
// Deprecated: use github.com/faustbrian/go-idempotency/adapters/queue. This
// package remains supported for the longer of 180 days after successor
// availability and two subsequently published stable root-module minor
// releases.
package idempotencyqueue

import (
	"context"
	"time"

	"github.com/faustbrian/go-idempotency"
	canonical "github.com/faustbrian/go-idempotency/adapters/queue"
)

var (
	// ErrInProgress tells a broker to retry after another owner's lease.
	ErrInProgress = canonical.ErrInProgress
	// ErrConflict identifies reuse of a delivery key for a different payload.
	ErrConflict = canonical.ErrConflict
	// ErrTerminalFailure identifies a previously recorded permanent failure.
	ErrTerminalFailure = canonical.ErrTerminalFailure
)

// Message is satisfied by queue core.TaskMessage and similar deliveries.
type Message interface {
	Payload() []byte
}

// Handler processes one elected delivery owner.
type Handler func(context.Context, Message) error

// KeyFunc creates the consumer and delivery-scoped semantic key.
type KeyFunc func(context.Context, Message) (idempotency.Key, error)

// FingerprintFunc computes canonical payload identity.
type FingerprintFunc func(Message) (idempotency.Fingerprint, error)

// Options configures durable consumer ownership and cleanup.
type Options struct {
	Service           *idempotency.Service
	Lease             time.Duration
	TransitionTimeout time.Duration
	Key               KeyFunc
	Fingerprint       FingerprintFunc
}

// Middleware preserves the legacy queue adapter type identity.
type Middleware struct{ inner *canonical.Middleware }

// New validates options and constructs queue middleware.
func New(options Options) (*Middleware, error) {
	var key canonical.KeyFunc
	if options.Key != nil {
		key = func(ctx context.Context, message canonical.Message) (idempotency.Key, error) {
			return options.Key(ctx, message)
		}
	}
	var fingerprint canonical.FingerprintFunc
	if options.Fingerprint != nil {
		fingerprint = func(message canonical.Message) (idempotency.Fingerprint, error) {
			return options.Fingerprint(message)
		}
	}
	inner, err := canonical.New(canonical.Options{
		Service: options.Service, Lease: options.Lease,
		TransitionTimeout: options.TransitionTimeout,
		Key:               key, Fingerprint: fingerprint,
	})
	if err != nil {
		return nil, err
	}

	return &Middleware{inner: inner}, nil
}

// Handle executes handler for an acquired delivery and completes it on success.
func (middleware *Middleware) Handle(
	ctx context.Context,
	message Message,
	handler Handler,
) error {
	var next canonical.Handler
	if handler != nil {
		next = func(ctx context.Context, message canonical.Message) error {
			return handler(ctx, message)
		}
	}

	return middleware.inner.Handle(ctx, message, next)
}

// Wrap preserves the concrete message type expected by queue WithFn.
func Wrap[M Message](
	middleware *Middleware,
	next func(context.Context, M) error,
) func(context.Context, M) error {
	return func(ctx context.Context, message M) error {
		return middleware.Handle(ctx, message, func(ctx context.Context, _ Message) error {
			return next(ctx, message)
		})
	}
}
