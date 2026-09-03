# Changelog

All notable changes to this project are documented in this file. The format is
based on Keep a Changelog, and releases follow Semantic Versioning after the
public API reaches its first stable version.

## [Unreleased]

### Changed

- Adopt `go-telemetry` v1.1.1 in the ecosystem compatibility harness so its
  OTLP graph selects the patched gRPC v1.83.1 release.
- Replace bootstrap-only Golib module hashes with their immutable public
  SumDB identities in the root and ecosystem compatibility modules.
- Adopt checksum-verified `go-library-tools` v1.4.0 enforcement so dependency
  resolution rejects bootstrap-only checksums before repository gates run.
- Advance consumer navigation to the v1.4.0 ecosystem index and its
  persistence-and-durability family guidance.

- Publish complete schema-v2 cohesion metadata and versioned ecosystem
  navigation for the stable idempotency module and its interoperability
  harness.
- Adopt the checksum-verified `go-library-tools` v1.3.0 CLI, add the local
  cohesion gate, and pin hosted cohesion enforcement to its immutable source
  revision.
- Replace the stale development-state summary with the module's released,
  stable workload coverage.

- Adopt the checksum-verified `go-library-tools` v1.2.0 CLI and immutable
  shared workflow so local and hosted gates enforce specification governance
  while preserving the API baseline, mutation checkpoints, fuzz targets,
  benchmarks, and package fixtures.

### Documentation

- Record RFC 9110 Erratum 9162 as behavior-neutral for the HTTP adapter, with
  no decision or conformance-binding changes.

- Add the [specification decision register](docs/specification-decisions.md),
  pinned source authorities, drift monitoring, and executable conformance
  records for the package's JCS, JSON-RPC, HTTP, digest, and durable-ownership
  policies.
- IDEMPOTENCY-DEC-001 sha256:f70e8d1c3f8b42a0fd7a3e0f8f953105b872d902878f5d5238b445a475675b03
- IDEMPOTENCY-DEC-002 sha256:038eaa1aab6c56e254724aa55b38ca11043d2d58093e764557af4def3d1ebf4f
- IDEMPOTENCY-DEC-003 sha256:9da3530377f546b5894f6165940839dd83cc2b6421db3513c1f36c5a8352a4db
- IDEMPOTENCY-DEC-004 sha256:752969969e6a47878bf97da1ed660f85c485671a86f5ed362020fa719bbd0d39
- IDEMPOTENCY-DEC-005 sha256:84157c145b9b72c4c6d705883440ee3cf564e765aa1b8a45d47702af33c72a2a
- IDEMPOTENCY-DEC-006 sha256:98a9dc305c8123885195d3084636a4da4e81618ee65def81e933586c6dae9d2e
- IDEMPOTENCY-DEC-007 sha256:c194ad1a0de1bb5ae63dd1fd9b4f49569c62a846f1acb1251fb44030a61f37b3
- IDEMPOTENCY-DEC-008 sha256:9430fa7e72a4e9ddb8372a8bf5a78d17575087f869ae1c42a016963f746405f6
- IDEMPOTENCY-DEC-009 sha256:1ea582419cf1aba0c888d5ceb6162629ae3fc9dec6faf05bb9f52d07beb7e5d2
- IDEMPOTENCY-DEC-010 sha256:c024926fbfbf6a656561ea01ece8d707703f509ec7b187e627b2aebe8b771e24
- IDEMPOTENCY-DEC-011 sha256:b1a3b56cf9ca5a1691b33fa253ed4b9f447782ae7f773bd1991c870efa92071b
- IDEMPOTENCY-DEC-012 sha256:1dce641d54df5f9bf861e23e9c6a2bf8216bb0edaa207c72198fb4f720d4f5c9
- IDEMPOTENCY-DEC-013 sha256:31493a3a629b4f4679cb9d2275fbe2d7f84e81c0ad95ad42a535ea00e1fc359b
- Remove completed implementation plans from the release tree and retain
  package-owned documentation as the maintained reference.

## [1.0.0] - 2026-08-25

### Compatibility

- Regenerate the exported API baseline with the repository's Go 1.26
  toolchain so JSON-backed contracts retain their intended stable identity.

### Changed

- Exclude intentional nested modules from root local-proxy archives so local,
  bootstrap, CI, and public module checksums describe the same source
  boundary.

- Track the pinned documentation-tool lockfile so clean CI checkouts install
  the exact validated cspell dependency.

- Reconcile standalone dependency checksums against deterministic current
  module archives so CI, local verification, and release consumers resolve
  identical content.

- Harden standalone documentation validation with deterministic spelling and
  link checks, package-specific documentation gates, and repository-local
  contributor guidance.

### Documentation

- Replace obsolete standalone-repository links and workflow claims with
  monorepo-canonical targets and current release guidance.

### Compatibility

- Added a pinned module export baseline so incompatible public API changes
  fail the canonical repository gate.

### Changed

- Publish the module from its standalone `github.com/faustbrian/go-idempotency` identity while preserving its documented API and behavior.
- Refresh local `v0.0.0` owned-module checksums after dependency manifests and
  release notes were normalized; runtime behavior and public APIs are
  unchanged.
- Update the webhook ecosystem contract and adoption guide to the canonical
  `webhook/adapters/idempotency` package path.
- Refresh the ecosystem compatibility dependency graph for the patched gRPC
  release selected by telemetry.
- Pinned unpublished owned `clock` and `migrations` dependencies to resolvable
  main-branch pseudo-versions so external consumers can install this module.
- Refresh owned-module checksums against the final consolidated archives.
- Kept the ecosystem compile-contract assertion explicit without redundantly
  spelling a generic function type that strict static analysis can infer.
- Kept typed-wrapper and fault-injection tests compatible with the canonical
  strict static-analysis configuration without weakening their assertions.
- Normalized standalone module metadata against the canonical owned dependency
  graph, including complete checksums for clean consumer resolution.

### Added

- Public PostgreSQL record-key digest derivation for business transactions that
  must lock and validate the same idempotency row before fenced side effects.
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
- Deterministic Valkey response-loss injection that targets the scripted write
  boundary after script warm-up, proving unknown-result recovery without
  accidentally dropping setup or discovery traffic.
- Bounded JSON canonicalization and byte fingerprint helpers.
- HTTP response replay, method-aware JSON-RPC result and error replay, queue and
  webhook delivery deduplication, and named command and import execution.
- Bounded, cancellation-independent panic cleanup across HTTP, JSON-RPC, queue,
  command, and webhook handler integrations.
- Fencing ownership propagation through handler contexts.
- Bounded service observations and keyed HMAC correlation without raw logical
  key exposure or high-cardinality metric fields.
- Typed `log`/`slog` and `telemetry`/OpenTelemetry observers for bounded
  transition logs and metrics.
- `outbox` transaction coordination that inserts an envelope and completes
  idempotency through the same caller-owned PostgreSQL transaction.
- Direct `migrations` schema binding and compatibility coverage for the
  `webhook` durable replay-store adapter.
- A pinned compatibility module covering the published `log`,
  `migrations`, `outbox`, `queue`, `telemetry`, and `webhook`
  contracts.
- Frozen PostgreSQL and Valkey version-1 record fixtures that lock retained
  reader and writer compatibility across rolling releases.
- Race, fuzz smoke, vulnerability, exact coverage, benchmark, and backend matrix
  automation.
- Exhaustive illegal-transition, stale-owner, duplicate-completion, crash-point,
  and fenced-resource proof suites shared by every backend.
- PostgreSQL failure injection for deadlocks, serializable aborts, pool
  saturation, rollback, response loss, and cleanup contention.
- Valkey 9 replica-promotion failure injection in local, CI, and release gates.
- Bounds for fingerprint policy versions and owner tokens, plus configurable
  bounded memory-store retention with a safe default.
- Hostile-input fuzz coverage for canonical JSON, duplicate object keys,
  Unicode forms, numeric forms, binary encodings, oversized input, and
  cross-version fingerprint identity.
- Formal threat model, hardening findings, resource budgets, crash and
  transition evidence, recovery obligations, and benchmark baselines.
- Five-minute adoption, concepts, operations, capacity, troubleshooting,
  migration, compatibility, security, contribution, and FAQ documentation.
- Semantic-version tag verification and least-privilege GitHub release
  automation.

### Known limitations

- The public API follows stable v1 semantic-versioning compatibility.

[Unreleased]: https://github.com/faustbrian/go-idempotency/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/faustbrian/go-idempotency/releases/tag/v1.0.0
