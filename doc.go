// Package idempotency defines durable ownership and replay semantics for
// operations that may be retried.
//
// The package does not guarantee exactly-once execution. A lease proves only
// that an owner may act during a bounded interval; it cannot prove that an
// expired owner stopped. Applications must carry the fencing token into the
// transaction or conditional write that protects the business side effect.
// Use NewKey/NewFingerprint and a provider-backed NewService; the core owns
// no backend closure, retry/wait loop, automatic heartbeat or transaction with
// external effects. Calls and instrumentation are synchronous; dependencies
// must honor non-nil contexts and support the chosen concurrent workload.
package idempotency
