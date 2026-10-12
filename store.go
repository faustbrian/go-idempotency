package idempotency

import (
	"context"
	"time"
)

// State is the durable lifecycle state of an idempotency record.
type State string

const (
	// StateAcquired means an owner holds a fresh lease but has not heartbeated.
	StateAcquired State = "acquired"
	// StateRunning means the current owner has extended its lease.
	StateRunning State = "running"
	// StateCompleted means a successful bounded result is terminal and replayable.
	StateCompleted State = "completed"
	// StateFailed means a bounded terminal failure is replayable.
	StateFailed State = "failed"
	// StateExpired means an elapsed active lease was explicitly recorded.
	StateExpired State = "expired"
	// StateAbandoned means the owner deliberately released without a result.
	StateAbandoned State = "abandoned"
)

// Outcome describes the semantic result of attempting acquisition.
type Outcome string

const (
	// OutcomeAcquired means the caller owns a new executable attempt.
	OutcomeAcquired Outcome = "acquired"
	// OutcomeReplayed means the same fingerprint has a completed result.
	OutcomeReplayed Outcome = "replayed"
	// OutcomeInProgress means another unexpired owner is current.
	OutcomeInProgress Outcome = "in_progress"
	// OutcomeConflict means the retained key has a different fingerprint.
	OutcomeConflict Outcome = "conflict"
	// OutcomeUnavailable means durable ownership could not be established.
	OutcomeUnavailable Outcome = "unavailable"
	// OutcomeStaleOwnerTakeover means the caller replaced an elapsed active owner.
	OutcomeStaleOwnerTakeover Outcome = "stale_owner_takeover"
	// OutcomeTerminalFailure means a terminal failure is available for replay.
	OutcomeTerminalFailure Outcome = "terminal_failure"
)

// Record is an exported snapshot of a retained state machine. Owned stores copy
// Result/Metadata on outward views; manually assembled records remain mutable.
// Lease expiry is not process termination, and purge ends the fencing domain.
type Record struct {
	// Key is the scoped logical identity; use NewKey before store admission.
	Key Key
	// Fingerprint binds the retained key to a versioned input digest.
	Fingerprint Fingerprint
	// State is the retained transition state, not proof external effects committed.
	State State
	// OwnerToken is opaque current-attempt proof; do not log or expose it.
	OwnerToken string
	// FencingToken increases within a retained record; deletion can reset its fencing
	// domain.
	FencingToken uint64
	// LeaseExpiresAt is exclusive backend-clock ownership expiry, not process termination.
	LeaseExpiresAt time.Time
	// HeartbeatAt is the last lease grant/renewal time, initially acquisition time.
	HeartbeatAt time.Time
	// Attempt counts acquired attempts within the retained record.
	Attempt uint64
	// CreatedAt is the first creation time of this retained generation.
	CreatedAt time.Time
	// UpdatedAt is the latest persisted transition time.
	UpdatedAt time.Time
	// CompletedAt is populated by successful completion.
	CompletedAt time.Time
	// FailedAt is populated by terminal failure persistence.
	FailedAt time.Time
	// AbandonedAt is populated by explicit live-owner release.
	AbandonedAt time.Time
	// ExpiredAt is populated by explicit expiry, not every elapsed lease.
	ExpiredAt time.Time
	// Result is copied replay bytes in owned store snapshots; application code owns
	// encryption and redaction.
	Result []byte
	// Metadata is copied stored diagnostic text; limits do not authenticate or redact it.
	Metadata map[string]string
}

// Ownership returns the proof required for current-owner mutations.
func (r Record) Ownership() Ownership {
	return Ownership{
		Key:          r.Key,
		OwnerToken:   r.OwnerToken,
		FencingToken: r.FencingToken,
	}
}

// Ownership identifies one leased attempt with an opaque token and fence.
type Ownership struct {
	// Key selects the record generation whose proof is being presented.
	Key Key
	// OwnerToken is the opaque elected token, not caller authentication.
	OwnerToken string
	// FencingToken must match current ownership; downstream effects need their own fence
	// enforcement.
	FencingToken uint64
}

// AcquireRequest supplies the stable identity, fingerprint, and requested lease.
type AcquireRequest struct {
	// Key is constructed scoped identity, not inferred from transport details.
	Key Key
	// Fingerprint must match every retry for the retained key, including terminal states.
	Fingerprint Fingerprint
	// Lease requests positive ownership duration at most MaxLease in provider time.
	Lease time.Duration
}

// AcquireResult contains the acquisition outcome and authoritative record.
type AcquireResult struct {
	// Outcome distinguishes executable ownership from retained replay/conflict/in-progress.
	Outcome Outcome
	// Record is a copied snapshot only authoritative when the operation succeeded.
	Record Record
}

// HeartbeatRequest extends a live current owner's lease.
type HeartbeatRequest struct {
	// Ownership must match a live current attempt.
	Ownership Ownership
	// Lease replaces expiry with now plus Lease, not an increment of old expiry.
	Lease time.Duration
}

// CompleteRequest records a bounded successful terminal result.
type CompleteRequest struct {
	// Ownership must match a live current attempt.
	Ownership Ownership
	// Result is copied bounded replay data; empty is valid.
	Result []byte
	// Metadata is copied bounded metadata; nil is valid.
	Metadata map[string]string
}

// FailRequest records a bounded terminal failure result.
type FailRequest struct {
	// Ownership must match a live current attempt.
	Ownership Ownership
	// Result is copied terminal-failure replay data, not retry instructions.
	Result []byte
	// Metadata is copied bounded metadata; nil is valid.
	Metadata map[string]string
}

// Store provides atomic retained-record transitions for valid constructed keys
// and fingerprints. Implementations must honor non-nil contexts, support concurrent
// operations and copy outward replay data. Durable providers use server time;
// memory is process-local despite successful tracked outcomes. Retention deletion
// ends key/fence history. An error/timeout can follow a committed mutation; only
// a successful result is authority, not a partially populated returned record.
type Store interface {
	// Acquire atomically selects ownership or retained outcome; conflict precedes
	// replay/live ownership.
	Acquire(context.Context, AcquireRequest) (AcquireResult, error)
	// Inspect returns retained state, which can have elapsed lease expiry; it does not
	// renew ownership.
	Inspect(context.Context, Key) (Record, error)
	// Heartbeat extends only matching live ownership in provider-authoritative time.
	Heartbeat(context.Context, HeartbeatRequest) (Record, error)
	// Complete conditionally stores replay data; an error does not prove an external
	// effect rolled back.
	Complete(context.Context, CompleteRequest) (Record, error)
	// Fail conditionally stores terminal failure, not a recoverable retry classification.
	Fail(context.Context, FailRequest) (Record, error)
	// Release abandons a matching live attempt without stopping its process.
	Release(context.Context, Ownership) (Record, error)
	// Expire records elapsed active ownership without granting a new owner or deleting
	// history.
	Expire(context.Context, Key) (Record, error)
}
