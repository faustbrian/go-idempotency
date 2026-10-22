# Compatibility Policy

The root module follows semantic versioning with root `v<version>` tags. Its
v3 API uses the `/v3` module and import suffix without a version-specific
source directory. The nested ecosystem module is an independent, non-releasable
compatibility harness and retains its own module path.

Before `v1`, minor releases MAY contain reviewed breaking changes, but every
break MUST be documented with migration guidance. Patch releases MUST remain
backward compatible. At and after `v1`, incompatible exported API or documented
behavior changes require a new major version.

The `postgres.GoMigration` result uses the `go-migrations/v4` named type in the
v3 module. Update the idempotency import path and migrations import path
together; earlier majors' values are distinct types. Moving from go-migrations
v2 to v4 also requires aligning the migration ledger table as described in the
[go-migrations ledger name upgrade](https://github.com/faustbrian/go-migrations/blob/v4.0.0/docs/operations.md#ledger-name-upgrade); the migration's SQL and checksum are unchanged.

Compatibility includes exported Go APIs, error classification, serialization,
protocol behavior, persistence schemas, environment variables, command output,
resource ownership, ordering, retry/idempotency semantics, and documented
defaults. A compile-compatible change can still be behaviorally breaking.

Specification-backed modules MUST NOT diverge from their declared standards.
Ambiguities require documented decisions and stable tests. Deprecated APIs
follow [`DEPRECATION.md`](DEPRECATION.md).

The [specification decision register](docs/specification-decisions.md) is part
of this compatibility contract. A canonicalization, fingerprint, projection,
status, replay, or durable-enforcement decision change requires compatibility
and changelog review even when prior behavior was undocumented.
