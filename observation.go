package idempotency

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

const minCorrelationSecretBytes = 32

// KeyHasher is trusted application correlation policy. Use NewHMACKeyHasher for
// keyed digest output; arbitrary callbacks can return raw or unbounded data.
type KeyHasher func(Key) string

// Transition identifies one fixed-cardinality semantic service operation.
type Transition string

const (
	// TransitionAcquire identifies a Service.Begin acquisition attempt.
	TransitionAcquire Transition = "acquire"
	// TransitionInspect identifies a Service.Inspect read.
	TransitionInspect Transition = "inspect"
	// TransitionHeartbeat identifies a Service.Heartbeat lease extension.
	TransitionHeartbeat Transition = "heartbeat"
	// TransitionComplete identifies a Service.Complete terminal transition.
	TransitionComplete Transition = "complete"
	// TransitionFail identifies a Service.Fail terminal transition.
	TransitionFail Transition = "fail"
	// TransitionRelease identifies a Service.Release abandonment transition.
	TransitionRelease Transition = "release"
	// TransitionExpire identifies a Service.Expire audit transition.
	TransitionExpire Transition = "expire"
)

// Observation is structured service instrumentation. Core transitions/outcomes
// are classified, but custom store reasons and KeyHasher output are not bounded
// or redacted. Manually constructed observations are arbitrary application data.
// Correlation is suitable for restricted logs, never metric labels.
type Observation struct {
	// Transition identifies the semantic operation.
	Transition Transition
	// Outcome is populated for acquisition attempts.
	Outcome Outcome
	// Reason classifies a failed transition and is empty on success.
	Reason Reason
	// Durable reports accepted tracked-store success, not a provider durability certificate.
	Durable bool
	// Correlation is trusted KeyHasher output; only the owned HMAC helper guarantees
	// digest shape.
	Correlation string
}

// Observer receives structured semantic service transition signals. Service
// isolates observer and key-hasher panics so instrumentation cannot change a
// semantic result. Implementations should still return quickly and honor ctx.
type Observer interface {
	// Observe is synchronous and may run concurrently; cooperation/latency are caller-owned.
	Observe(context.Context, Observation)
}

// ObserverFunc adapts a function to Observer.
type ObserverFunc func(context.Context, Observation)

// Observe calls f synchronously with the supplied observation; direct calls
// have no panic containment or input normalization.
func (f ObserverFunc) Observe(ctx context.Context, observation Observation) {
	f(ctx, observation)
}

// ServiceOptions configures optional synchronous trusted instrumentation.
type ServiceOptions struct {
	// Observer receives attempted delivery after instrumented transitions; hasher panics
	// can omit it.
	Observer Observer
	// KeyHasher is trusted correlation policy; use the HMAC helper rather than returning
	// raw keys.
	KeyHasher KeyHasher
}

// NewHMACKeyHasher constructs a deterministic SHA-256 HMAC key hasher.
// The secret is copied and must contain at least 32 bytes; it is retained in
// the returned closure until unreachable, with no erasure/Close guarantee.
// Deterministic correlation is not anonymization when an attacker knows the secret.
func NewHMACKeyHasher(secret []byte) (KeyHasher, error) {
	if len(secret) < minCorrelationSecretBytes {
		return nil, &Error{
			Reason: ReasonInvalidConfiguration,
			Field:  "correlation_secret",
		}
	}
	copied := append([]byte(nil), secret...)
	return func(key Key) string {
		digest := hmac.New(sha256.New, copied)
		var size [4]byte
		for _, part := range []string{
			key.Namespace(), key.Tenant(), key.Operation(), key.Caller(), key.Value(),
		} {
			// #nosec G115 -- private Key parts are empty or bounded to 256 bytes by NewKey.
			binary.BigEndian.PutUint32(size[:], uint32(len(part)))
			_, _ = digest.Write(size[:])
			_, _ = digest.Write([]byte(part))
		}
		return hex.EncodeToString(digest.Sum(nil))
	}, nil
}
