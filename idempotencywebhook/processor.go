// Package idempotencywebhook is the legacy durable webhook adapter.
//
// Deprecated: use github.com/faustbrian/go-idempotency/v2/adapters/webhook. This
// package remains supported for the longer of 180 days after successor
// availability and two subsequently published stable root-module minor
// releases.
package idempotencywebhook

import (
	"context"
	"time"

	"github.com/faustbrian/go-idempotency/v2"
	canonical "github.com/faustbrian/go-idempotency/v2/adapters/webhook"
)

var (
	// ErrInProgress reports another active delivery owner.
	ErrInProgress = canonical.ErrInProgress
	// ErrConflict reports reuse of a provider delivery ID for another payload.
	ErrConflict = canonical.ErrConflict
	// ErrTerminalFailure reports a deliberately persisted permanent failure.
	ErrTerminalFailure = canonical.ErrTerminalFailure
)

// Delivery is the structural payload contract expected from webhook messages.
type Delivery interface {
	Payload() []byte
}

// Handler processes one elected provider delivery.
type Handler func(context.Context, Delivery) error

// KeyFunc constructs a provider, endpoint, tenant, and delivery-scoped key.
type KeyFunc func(context.Context, Delivery) (idempotency.Key, error)

// FingerprintFunc computes canonical event identity.
type FingerprintFunc func(Delivery) (idempotency.Fingerprint, error)

// Options configures durable webhook delivery ownership.
type Options struct {
	Service           *idempotency.Service
	Lease             time.Duration
	TransitionTimeout time.Duration
	Key               KeyFunc
	Fingerprint       FingerprintFunc
}

// Processor preserves the legacy webhook adapter type identity.
type Processor struct{ inner *canonical.Processor }

// New validates options and constructs a webhook processor.
func New(options Options) (*Processor, error) {
	var key canonical.KeyFunc
	if options.Key != nil {
		key = func(ctx context.Context, delivery canonical.Delivery) (idempotency.Key, error) {
			return options.Key(ctx, delivery)
		}
	}
	var fingerprint canonical.FingerprintFunc
	if options.Fingerprint != nil {
		fingerprint = func(delivery canonical.Delivery) (idempotency.Fingerprint, error) {
			return options.Fingerprint(delivery)
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

	return &Processor{inner: inner}, nil
}

// Handle executes handler once and deduplicates completed redelivery.
func (processor *Processor) Handle(
	ctx context.Context,
	delivery Delivery,
	handler Handler,
) error {
	var next canonical.Handler
	if handler != nil {
		next = func(ctx context.Context, delivery canonical.Delivery) error {
			return handler(ctx, delivery)
		}
	}

	return processor.inner.Handle(ctx, delivery, next)
}

// Wrap preserves the provider-specific delivery type used by a webhook router.
func Wrap[D Delivery](
	processor *Processor,
	next func(context.Context, D) error,
) func(context.Context, D) error {
	if next == nil {
		return func(context.Context, D) error { return configurationError("handler") }
	}

	return func(ctx context.Context, delivery D) error {
		return processor.Handle(ctx, delivery, func(ctx context.Context, _ Delivery) error {
			return next(ctx, delivery)
		})
	}
}

func configurationError(field string) error {
	return &idempotency.Error{Reason: idempotency.ReasonInvalidConfiguration, Field: field}
}
