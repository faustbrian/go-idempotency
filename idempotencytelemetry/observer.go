// Package idempotencytelemetry is the legacy OpenTelemetry adapter.
//
// Deprecated: use github.com/faustbrian/go-idempotency/adapters/otel. This
// package remains supported for the longer of 180 days after successor
// availability and two subsequently published stable root-module minor
// releases.
package idempotencytelemetry

import (
	"context"

	"github.com/faustbrian/go-idempotency"
	canonical "github.com/faustbrian/go-idempotency/adapters/otel"
	"go.opentelemetry.io/otel/metric"
)

// ErrNilMeterProvider reports an unusable telemetry configuration.
var ErrNilMeterProvider = canonical.ErrNilMeterProvider

// Observer preserves the legacy OpenTelemetry adapter type identity.
type Observer struct{ inner *canonical.Observer }

// New constructs an observer from a standard OpenTelemetry meter provider.
func New(provider metric.MeterProvider) (*Observer, error) {
	inner, err := canonical.New(provider)
	if err != nil {
		return nil, err
	}

	return &Observer{inner: inner}, nil
}

// Observe increments the transition counter.
func (observer *Observer) Observe(ctx context.Context, event idempotency.Observation) {
	observer.inner.Observe(ctx, event)
}

var _ idempotency.Observer = (*Observer)(nil)
