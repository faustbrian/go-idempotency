# Migrations and compatibility

The module is pre-v1. Source compatibility may change between minor releases
until a stable API is declared, but persisted records and rolling deployments
still require explicit compatibility discipline.

## Supported runtime matrix

The current development line requires Go 1.25 or later, PostgreSQL 16 or 17
for the PostgreSQL adapter, and Valkey 9 or later for the Valkey adapter.
Standalone and three-primary Valkey Cluster topologies are tested. A topology
claim covers only the adapter's one-key scripts; it does not make application
multi-key work atomic.

The release notes must identify any change to these minimums. CI is the source
of truth for the versions tested by each commit.

## Public Go API policy

Before v1, minor releases may remove or change exported identifiers. Patch
releases should remain source-compatible unless a security or correctness flaw
makes that unsafe. Once v1 is released, incompatible public API changes require
a new major version under Go module semantic import versioning.

Applications should pin a tagged version, review `CHANGELOG.md`, compile all
integration packages they use, and run their own crash and retry tests before
upgrading. Do not deploy an unreviewed `main` snapshot as a production upgrade.

## Persisted record policy

PostgreSQL JSONB records and Valkey hashes contain an explicit schema version.
Version `1` is the only current version. Decoders reject unknown versions and
malformed semantic fields; they do not guess or partially recover a record.

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

`postgres.SchemaMigration()` returns the reversible version-1 table migration.
Apply it through deployment tooling before application readiness. The
`go-migrations` repository currently has no public module or binding API, so
the package exposes a neutral descriptor rather than inventing an integration.

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

`go-log` exposes `*slog.Logger`, and `go-telemetry.Runtime` exposes standard
OpenTelemetry providers. The `idempotencylog` and `idempotencytelemetry`
packages bind those contracts without depending on unreleased upstream commit
identities. `go-queue/core.TaskMessage` satisfies the structural queue message
contract.

`go-migrations` and `go-webhook` currently have no public Go module APIs, while
`go-outbox` has no published Go module or record API. The neutral migration
descriptor, structural webhook integration, and transaction/outbox recipe
avoid claiming compatibility with contracts that do not exist. Future direct
bindings must be additive, version-pinned once tags exist, and tested against
the upstream module's supported versions.
