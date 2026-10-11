# Contributing

## Before Editing

1. Read [`AGENTS.md`](AGENTS.md) and the affected package's docs.
2. Run the affected package's deterministic Go tests as described below.
3. Identify owned dependencies and reverse dependants in `modules.json`.
4. Preserve unrelated work and generated/corpus provenance.

## Changes

Keep commits focused and conventional. Update every affected changelog with
the behavior and migration impact. Public API changes require compatibility
evidence and documentation. Specification behavior requires a decision record,
fixture coverage, and interoperability evidence.

New direct dependencies and dependency updates must follow the
[dependency governance policy](AGENTS.md#dependencies-and-supply-chain). Package-local
update bots are forbidden; the root policy owns every module and action update.

Repository verification is provided by the released `go-library-tools` version
and checksum in `.golib.yaml`. The Makefile is a thin local entry point; the
API baseline and mutation checkpoints remain repository-owned.

Specification-backed changes must follow the
[specification governance contract](AGENTS.md#design), update
the affected stable decision entries, and complete the Specification Decisions
section of the pull request template. An unresolved interpretation or stale
source pin is release-blocking; peer behavior cannot silently select policy.
The [specification decision register](docs/specification-decisions.md) and its
machine bindings preserve superseded entries and append a digest whenever a
recorded decision changes.

Selected full/release mutation gates must finish with zero surviving viable
mutants; mutation is not required for an ordinary local or pull-request check.

Do not add package-local workflows, permanent replacements, machine-specific
paths, bypass flags, broad mutation exclusions, or aggregate quality metrics
that hide a failing package.

## Verification

Use Go 1.27.2 for the producer; both module/API minima remain Go 1.27.0.
An isolated checkout needs no sibling repositories or coordination goal files.
Start with deterministic tests and, when its hostile-input boundary changes,
one explicit active fuzz target:

```bash
GOWORK=off go test ./...
GOWORK=off go test ./canonical -run '^$' -fuzz '^FuzzJSONIsIdempotent$' -fuzztime=2s -timeout=2m
```

The first command includes fuzz seeds/corpus, not active fuzzing. Backend tests
skip when `POSTGRES_URL` or `VALKEY_ADDR` is absent; a direct unconfigured Go
test run does not establish backend qualification. Cluster/failover tests use
the additional backend configuration documented by their suites.

Install the verified public `golib` release declared in `.golib.yaml`; ordinary
CI verifies the same release without building it from source:

```bash
make check
make ci
make specification-check
```

`make check` and `make ci` retain PostgreSQL and Valkey service prerequisites for
both the root and intentional v1 `compatibility/ecosystem` module; no test tag
is declared. They are not database- or Docker-free. `ci` also validates
repository/cohesion structure; specification checks use the offline decision
register. The memory race runs once with a two-minute timeout and checks
single-owner election, fenced takeover, stale-owner refusal and retained
current results. Base-free local checks run deterministic fuzz seeds/corpus;
changed-source CI selects active smoke. Run the owned public-contract harness
from `compatibility/ecosystem` with `GOWORK=off go test ./...` when composition
changes; its v1 dependencies are intentional and are not migrated by this route.

Optional one-package local mutation:

```bash
golib mutation --module . --package memory
```

Select full qualification and release rehearsal explicitly:

```bash
make check-full
make conformance
make specification-check-online
golib api check
golib release dry-run --module .
```

Full checks retain all nine native fuzz targets at 10000x/20m each,
conformance/interoperability, benchmarks and original module gates.
Conformance installs the locked JavaScript peer for JCS differential checks;
online specification verification fetches pinned authority/errata sources.
Release rehearsal can run locally, not only in CI. Report unavailable or
failing commands and never describe bounded checks alone as release
qualification.

## Adding A Module

Follow [repository structure policy](AGENTS.md#repository-structure). New modules
require an explicit purpose, ownership boundary, dependency review, package
catalog entry, full quality gates, documentation, changelog, license, security
policy, compatibility plan, and release dry-run.
