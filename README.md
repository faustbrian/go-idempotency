# idempotency

[![CI](https://github.com/faustbrian/go-idempotency/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/faustbrian/go-idempotency/actions/workflows/ci.yml)
[![CodeQL](https://img.shields.io/badge/CodeQL-required-blue)](https://github.com/faustbrian/go-idempotency/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/badge/coverage-risk_selected-blue)](CONTRIBUTING.md#verification)
[![Mutation](https://img.shields.io/badge/mutation-optional_local%2Fselected_scheduled-blue)](CONTRIBUTING.md#verification)
[![Documentation](https://img.shields.io/badge/docs-checked_in_CI-blue)](docs/)
[![Go Reference](https://pkg.go.dev/badge/github.com/faustbrian/go-idempotency/v3.svg)](https://pkg.go.dev/github.com/faustbrian/go-idempotency/v3)
[![Release](https://img.shields.io/github/v/release/faustbrian/go-idempotency?sort=semver)](https://github.com/faustbrian/go-idempotency/releases)
[![Go](https://img.shields.io/badge/go-1.27.0-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

`idempotency` provides durable ownership, fencing, and bounded result replay
for retried Go operations across HTTP, JSON-RPC, webhook, queue, import, and
command workloads.

The package deliberately does **not** claim exactly-once execution. A lease can
expire while an old process is still performing a side effect. Correct callers
must use the returned fencing token in a database transaction, conditional
write, or another application invariant that rejects stale owners.

## Status

The public contract, deterministic memory adapter, bounded JSON
canonicalization, PostgreSQL and Valkey adapters, buffered HTTP middleware,
method-aware JSON-RPC middleware, queue and webhook deduplication, named command
and import helpers, transactional `outbox` coordination, and bounded logging
and telemetry observers are implemented. The stable `/v3` module is the
current major. Untagged `main` is not a supported installation. Supported
release lines are listed in the [security policy](SECURITY.md).

New integrations use target-oriented packages under `adapters/`:

| Concern | Canonical package |
| --- | --- |
| Commands and imports | `github.com/faustbrian/go-idempotency/v3/adapters/command` |
| HTTP | `github.com/faustbrian/go-idempotency/v3/adapters/http` |
| JSON-RPC | `github.com/faustbrian/go-idempotency/v3/adapters/jsonrpc` |
| OpenTelemetry | `github.com/faustbrian/go-idempotency/v3/adapters/otel` |
| Transactional outbox | `github.com/faustbrian/go-idempotency/v3/adapters/outbox` |
| Queue consumers | `github.com/faustbrian/go-idempotency/v3/adapters/queue` |
| Structured logging | `github.com/faustbrian/go-idempotency/v3/adapters/slog` |
| Webhooks | `github.com/faustbrian/go-idempotency/v3/adapters/webhook` |

Version 3 removed the former `idempotencycommand`, `idempotencyhttp`,
`idempotencylog`, `idempotencyoutbox`, `idempotencyqueue`, `idempotencyrpc`,
`idempotencytelemetry`, and `idempotencywebhook` facades; import the matching
`adapters/*` path. `postgres.GoMigration` returns the `go-migrations/v4` type.

## Core acquisition outcomes

- `acquired`: this caller owns the first or a deliberately released attempt.
- `stale_owner_takeover`: this caller owns a new fenced attempt after expiry.
- `in_progress`: another unexpired owner is current.
- `replayed`: the same fingerprint has a completed bounded result.
- `terminal_failure`: the same fingerprint has a recorded terminal failure.
- `conflict`: the key already identifies a different fingerprint.
- `unavailable`: ownership could not be established; execution must fail closed
  unless an integration exposes and the caller selects an explicit duplicate-
  tolerant policy.

## Start in five minutes

Install the published v3 module:

```sh
go get github.com/faustbrian/go-idempotency/v3
```

The [quickstart](docs/quickstart.md) demonstrates acquisition, completion, and
replay with the deterministic API, then routes production deployments to the
PostgreSQL or Valkey 9 adapter. Read the [comparison with locks, unique
constraints, retries, and exactly-once claims](docs/concepts.md) before
choosing the application's side-effect invariant.

See [the state machine](docs/state-machine.md) and
[the crash semantics](docs/crash-semantics.md) before using the package. The
[HTTP middleware guide](docs/http.md) covers bounded handler response replay,
and the [JSON-RPC guide](docs/json-rpc.md) covers result and protocol-error
replay. The [queue guide](docs/queue.md) explains broker settlement behavior.
The [command and import guide](docs/commands-and-imports.md) covers stable source
record identities.
The [PostgreSQL guide](docs/postgres.md) covers transactional locking, cleanup,
permissions, and persisted-record privacy.
The [transaction and outbox guide](docs/outbox.md) shows atomic business,
`outbox` envelope, and completion commits with `adapters/outbox`.
The [webhook guide](docs/webhooks.md) covers signature ordering, provider
delivery identities, and response mapping.
The [operations guide](docs/operations.md) covers health, observability,
retention, cleanup, capacity, and incident recovery. The [threat
model](docs/threat-model.md), [assurance review](docs/assurance.md), and
[resource budgets](docs/resource-budgets.md) define the verified security and
operational envelope. See
[bounded logging and telemetry integration](docs/observability.md),
[troubleshooting](docs/troubleshooting.md), [migration and compatibility
policy](docs/migrations-and-compatibility.md), and the [FAQ](docs/faq.md) for
production adoption.

Licensed under the [MIT License](LICENSE). Release history is maintained in the
[changelog](CHANGELOG.md), and the complete guide index is in
[docs/README.md](docs/README.md).

For development, use Go 1.27.2 and `GOWORK=off go test ./...`; absent backend
configuration causes integration tests to skip, not qualify those backends.
`make check`/`make ci` retain the declared PostgreSQL/Valkey services for both
catalog modules and a selected memory ownership race. Base-free checks run
fuzz seeds/corpus, not active campaigns. See
[contributing](CONTRIBUTING.md#verification) for explicit focused fuzzing and
`make check-full`/release qualification.

Material canonicalization, JSON-RPC, HTTP, and durable-enforcement choices are
recorded in the [specification decision register](docs/specification-decisions.md).
Shared construction, ownership, lifecycle, and composition expectations are in
the versioned [Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.6.1/docs/ecosystem/README.md)
and its [Persistence and durability family](https://github.com/faustbrian/go-library-tools/blob/v1.6.1/docs/ecosystem/design-language.md#package-families-and-selection).

## Documentation navigation

[Complete documentation map](docs/README.md#complete-documentation-map) · [Current source package map](docs/package-map.md)

Use the [versioned ecosystem entry](https://github.com/faustbrian/go-library-tools/blob/v1.6.1/docs/ecosystem/README.md) for package
selection. The [frozen compatibility cohort](https://github.com/faustbrian/go-library-tools/blob/v1.8.4/docs/ecosystem/compatibility-sets.md)
is separate from current package releases and unpublished source changes.
Adopting them does not silently update that cohort.
