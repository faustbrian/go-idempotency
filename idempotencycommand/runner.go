// Package idempotencycommand is the legacy command and import adapter.
//
// Deprecated: use github.com/faustbrian/go-idempotency/adapters/command. This
// package remains supported for the longer of 180 days after successor
// availability and two subsequently published stable root-module minor
// releases.
package idempotencycommand

import (
	"context"
	"time"

	"github.com/faustbrian/go-idempotency"
	canonical "github.com/faustbrian/go-idempotency/adapters/command"
)

var (
	// ErrInProgress reports an unexpired owner for the source record.
	ErrInProgress = canonical.ErrInProgress
	// ErrConflict reports reuse of a source identity for different input.
	ErrConflict = canonical.ErrConflict
	// ErrTerminalFailure reports a deliberately persisted permanent failure.
	ErrTerminalFailure = canonical.ErrTerminalFailure
)

// Request identifies a named command or one record in an import source.
type Request struct {
	Namespace   string
	Tenant      string
	Name        string
	Caller      string
	SourceID    string
	Fingerprint idempotency.Fingerprint
}

// Handler executes one elected source-record owner and returns replay data.
type Handler func(context.Context) ([]byte, map[string]string, error)

// Options configures command leases and failed-handler cleanup.
type Options struct {
	Service           *idempotency.Service
	Lease             time.Duration
	TransitionTimeout time.Duration
}

// Result reports the semantic outcome and bounded replay data.
type Result struct {
	Outcome  idempotency.Outcome
	Result   []byte
	Metadata map[string]string
	Replayed bool
}

// Runner preserves the legacy command adapter type identity.
type Runner struct{ inner *canonical.Runner }

// New validates options and constructs a command runner.
func New(options Options) (*Runner, error) {
	inner, err := canonical.New(canonical.Options{
		Service: options.Service, Lease: options.Lease,
		TransitionTimeout: options.TransitionTimeout,
	})
	if err != nil {
		return nil, err
	}

	return &Runner{inner: inner}, nil
}

// Run executes handler only for a newly acquired or taken-over source record.
func (runner *Runner) Run(
	ctx context.Context,
	request Request,
	handler Handler,
) (Result, error) {
	result, err := runner.inner.Run(ctx, canonical.Request(request), canonical.Handler(handler))

	return Result(result), err
}
