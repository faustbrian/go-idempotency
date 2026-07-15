# Changelog

All notable changes to this project are documented in this file. The format is
based on Keep a Changelog, and releases follow Semantic Versioning after the
public API reaches its first stable version.

## [Unreleased]

### Added

- Durable semantic core with namespaced keys, canonical fingerprints, owner and
  fencing tokens, leases, heartbeats, attempts, terminal results, typed errors,
  and explicit acquisition outcomes.
- Deterministic in-memory adapter and shared store conformance suite.
- PostgreSQL adapter with advisory and row locking, server-clock leases,
  versioned JSONB records, bounded cleanup, transaction-bound completion, native
  fault tests, and PostgreSQL 16 and 17 integration coverage.
- Valkey 9 adapter with native `valkey-go`, atomic scripts, opaque cluster-safe
  keys, server-clock leases, explicit TTLs, startup safety checks, unknown-result
  recovery tests, and standalone and three-primary cluster coverage.
- Bounded JSON canonicalization and byte fingerprint helpers.
- HTTP response replay, method-aware JSON-RPC result and error replay, queue and
  webhook delivery deduplication, and named command and import execution.
- Fencing ownership propagation through handler contexts.
- Bounded service observations and keyed HMAC correlation without raw logical
  key exposure or high-cardinality metric fields.
- Typed `go-log`/`slog` and `go-telemetry`/OpenTelemetry observers for bounded
  transition logs and metrics.
- Race, fuzz smoke, vulnerability, exact coverage, benchmark, and backend matrix
  automation.
- Five-minute adoption, concepts, operations, capacity, troubleshooting,
  migration, compatibility, security, contribution, and FAQ documentation.
- Semantic-version tag verification and least-privilege GitHub release
  automation.

### Known limitations

- The public API is pre-v1 and may change before the first stable release.
- Direct typed bindings for `go-migrations` and `go-webhook` await public module
  APIs. `go-outbox` has no published Go module or record API to bind yet.

[Unreleased]: https://github.com/faustbrian/go-idempotency/commits/main
