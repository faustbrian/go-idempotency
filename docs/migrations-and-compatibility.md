# Migrations and compatibility

[Documentation index](README.md) · [Repository entry](../README.md)

The v3 module changes the published `postgres.GoMigration` result to the
`go-migrations/v4` named type and removes the v1.1 compatibility facades. V2
stays available at its original import path; consumers must migrate the
idempotency and migrations imports together. Persisted records and rolling
deployments still require explicit compatibility discipline.

## Supported runtime matrix

The current development line requires Go 1.27.0 or later, PostgreSQL 16 or 17
for the PostgreSQL adapter, and Valkey 9 or later for the Valkey adapter.
Standalone and three-primary Valkey Cluster topologies are tested. A topology
claim covers only the adapter's one-key scripts; it does not make application
multi-key work atomic.

The release notes must identify any change to these minimums. CI is the source
of truth for the versions tested by each commit.

Version 1.2 raises the Go floor from 1.26.6 to 1.27.0. Upgrade the consumer
toolchain before selecting v1.2; no module-path or import-path migration is
required. The release also makes Valkey schema-v1 validation atomic with each
script. Records written by this module remain compatible, while malformed,
oversized, or logically mis-keyed hashes fail closed before mutation.

## Public Go API policy

Major releases may remove or change exported identifiers. Patch
releases should remain source-compatible unless a security or correctness flaw
makes that unsafe. Once v1 is released, incompatible public API changes require
a new major version under Go module semantic import versioning.

Applications should pin a tagged version, review `CHANGELOG.md`, compile all
integration packages they use, and run their own crash and retry tests before
upgrading. Do not deploy an unreviewed `main` snapshot as a production upgrade.

## Adapter path migration

Version 1.1 added target-oriented adapter paths and kept every earlier
integration path as a compatibility facade through v2. Version 3 removes the
facades; replace each import with its canonical path:

| Removed facade path | Canonical path |
| --- | --- |
| `idempotencycommand` | `adapters/command` |
| `idempotencyhttp` | `adapters/http` |
| `idempotencylog` | `adapters/slog` |
| `idempotencyoutbox` | `adapters/outbox` |
| `idempotencyqueue` | `adapters/queue` |
| `idempotencyrpc` | `adapters/jsonrpc` |
| `idempotencytelemetry` | `adapters/otel` |
| `idempotencywebhook` | `adapters/webhook` |

Public signatures, stable error sentinels, durable state transitions, payload
bounds, ownership, cancellation, and replay behavior of the canonical packages
match the removed facades. Code that named a facade's own concrete or named
type must switch to the canonical package's type. The facades remain available
and supported on the v2 line under the security policy.

## Persisted record policy

PostgreSQL JSONB records and Valkey hashes contain an explicit schema version.
Version `1` is the only current version. Decoders reject unknown versions and
malformed semantic fields; they do not guess or partially recover a record.
Frozen version-1 PostgreSQL JSONB and Valkey hash fixtures are decoded and
re-encoded in CI so changes to field names, timestamp representation, binary
encoding, or omission behavior cannot silently redefine retained records.
The shared conformance corpus also proves that an identical SHA-256 digest
under a different fingerprint policy version is a conflict, never a replay.

Every persisted-format change must:

1. assign a new schema version when interpretation changes;
2. keep readers for every version that can remain within retention;
3. include fixtures and malformed-record tests for old and new versions;
4. prove old-reader/new-writer behavior or prevent that deployment order;
5. define forward and rollback deployment sequences;
6. document storage and migration impact in the changelog;
7. preserve key identity, fingerprint meaning, ownership proofs, fence
   monotonicity, timestamps, bounds, and terminal replay semantics.

Never reuse a version number, reinterpret a field in place, reset a retained
fence, or silently coerce invalid data. A migration may add representation but
must not manufacture evidence that an operation completed.

## Rolling deployment sequence

For an additive format that old code can safely ignore:

1. deploy readers that understand both the old and new versions but still
   write the old version;
2. verify the entire fleet is compatible;
3. enable new-version writes;
4. wait at least the full retention and retry window before removing the old
   reader;
5. remove old compatibility code only in a later release.

If old readers must reject the new version, do not mix old readers with new
writers. Drain or replace the old fleet before enabling writes. Rollback then
requires disabling new writes while a compatible reader remains deployed; it
must not downgrade existing records destructively.

## PostgreSQL schema migration

`postgres.SchemaMigration()` returns the reversible version-1 descriptor for
deployment tooling. `postgres.GoMigration()` returns the same validated SQL as
an immutable `migrations.Migration`, including its default transactional
mode and content checksum. Add that value to the `migrations` source used by
the deployment runner and apply it before application readiness.

The migration's version, name, SQL, and checksum are unchanged from v2. The
go-migrations v4 runner records history in `public.migrations` rather than the
v2 `public.go_schema_migrations` ledger, so stop older runners and align the
existing ledger rows as the
[go-migrations ledger name upgrade](https://github.com/faustbrian/go-migrations/blob/v4.0.0/docs/operations.md#ledger-name-upgrade) describes before
the first v4 run.

Future SQL migrations must tolerate the supported rolling sequence, avoid a
long blocking rewrite on a populated table, and include an operational plan
for index creation, backfill, validation, rollback, and cleanup. Test both an
empty database and a representative retained dataset.

Dropping the table or deleting records is not a normal rollback. It removes
ownership, replay, and fencing history and may authorize duplicate execution.

## Valkey format migration

Valkey records expire independently. A new reader must remain compatible for
at least the maximum configured retention across the fleet. Do not migrate by
renaming prefixes without treating the new prefix as a new fencing domain: the
same logical request would appear missing there.

An online rewrite must preserve TTL conservatively, use an atomic operation per
record, and handle concurrent transitions. Prefer read-old/write-current
decoders over bulk rewriting when possible. Never use eviction or flush
commands as a migration mechanism.

## Fingerprint and key migrations

Fingerprint versions identify canonical business policies. Changing the bytes
while retaining a version is a compatibility bug. Changing the version for the
same retained key produces a deliberate conflict, not automatic migration.

When a business policy must change, choose one of these explicitly:

- keep the old policy through its retry and retention window;
- introduce a new operation or key generation;
- reject old requests at the application boundary and reconcile them;
- perform an application-specific migration that proves the old and new
  requests are equivalent before changing identity.

Namespace, tenant, operation, caller, and value are all identity fields.
Changing any of them creates a distinct record and fencing domain.

## Ecosystem bindings

`log` exposes `*slog.Logger`, and `telemetry.Runtime` exposes standard
OpenTelemetry providers. The `adapters/slog` and `adapters/otel` packages bind
those contracts without depending on unreleased upstream commit identities.
`queue/core.TaskMessage` satisfies the structural queue message contract.
`adapters/outbox.InsertAndComplete` accepts the published
`outbox/postgres.Writer` and the idempotency PostgreSQL store through their
exact transaction method contracts.

The isolated `compatibility/ecosystem` module pins published v1 idempotency and
migrations contracts alongside `log`, Outbox `/v2` v2.0.0, `queue`, `telemetry`, and
`webhook` under Go 1.27. This remains an intentional v1 consumer while the
root module moves to `/v2`, because the published webhook replay adapter accepts
the v1 service type. It proves those exact public interfaces and runs the
webhook replay-store adapter against the deterministic memory store at every
CI and release gate. Pseudo-versions are commit pins, not stability claims;
review upstream changelogs before updating them.

The Outbox v2 PostgreSQL writer composes with the existing generic
`adapters/outbox.InsertAndComplete` boundary; its actual v2 envelope type is
compile-checked. This does not change transaction ownership or introduce an
Idempotency root release. Webhook selects public `/v3` v3.0.0 in this harness;
its replay adapter still accepts the intentional v1 Idempotency service.

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
