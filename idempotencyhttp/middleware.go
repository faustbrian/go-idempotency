package idempotencyhttp

import (
	"net/http"
	"time"

	"github.com/faustbrian/go-idempotency"
	canonical "github.com/faustbrian/go-idempotency/adapters/http"
)

const (
	// HeaderKey carries the caller's idempotency key.
	HeaderKey = canonical.HeaderKey
	// HeaderOutcome reports the semantic result of the request.
	HeaderOutcome = canonical.HeaderOutcome
	// HeaderReplayed is true when the response came from a durable record.
	HeaderReplayed = canonical.HeaderReplayed
	// MaxReplayResponseBytes is the largest handler body accepted for replay.
	MaxReplayResponseBytes = canonical.MaxReplayResponseBytes
)

// ErrResponseTooLarge is returned to a handler whose body crosses its limit.
var ErrResponseTooLarge = canonical.ErrResponseTooLarge

// KeyFunc maps a request and header value to an application-scoped key.
type KeyFunc func(*http.Request, string) (idempotency.Key, error)

// FingerprintFunc computes the application's canonical request fingerprint.
type FingerprintFunc func(*http.Request) (idempotency.Fingerprint, error)

// Options configures durable HTTP handler ownership and replay.
type Options struct {
	// Service owns the durable state machine.
	Service *idempotency.Service
	// Lease bounds one handler owner's authority.
	Lease time.Duration
	// MaxResponseBytes bounds the buffered handler body. Zero defaults to 64 KiB.
	MaxResponseBytes int
	// ReplayHeaders names response headers to persist and replay.
	ReplayHeaders []string
	// TransitionTimeout bounds detached panic cleanup. Zero defaults to five seconds.
	TransitionTimeout time.Duration
	// Key constructs the fully scoped semantic key.
	Key KeyFunc
	// Fingerprint supplies canonical business-request identity.
	Fingerprint FingerprintFunc
}

// Middleware preserves the legacy HTTP adapter type identity.
type Middleware struct{ inner *canonical.Middleware }

// New validates options and constructs middleware.
func New(options Options) (*Middleware, error) {
	inner, err := canonical.New(canonical.Options{
		Service: options.Service, Lease: options.Lease,
		MaxResponseBytes: options.MaxResponseBytes, ReplayHeaders: options.ReplayHeaders,
		TransitionTimeout: options.TransitionTimeout,
		Key:               canonical.KeyFunc(options.Key),
		Fingerprint:       canonical.FingerprintFunc(options.Fingerprint),
	})
	if err != nil {
		return nil, err
	}

	return &Middleware{inner: inner}, nil
}

// Handler wraps next with durable acquisition and response replay.
func (middleware *Middleware) Handler(next http.Handler) http.Handler {
	return middleware.inner.Handler(next)
}
