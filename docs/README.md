# Documentation

[Repository entry](../README.md) · [Versioned ecosystem](https://github.com/faustbrian/go-library-tools/blob/v1.6.1/docs/ecosystem/README.md)

- [Specification decisions](specification-decisions.md): authoritative source,
  conformance, compatibility, and reconsideration choices.

## Start here

- [Five-minute quickstart](quickstart.md)
- [State machine](state-machine.md)
- [Crash semantics](crash-semantics.md)
- [Idempotency and related mechanisms](concepts.md)
- [Fingerprint policies](fingerprints.md)
- [PostgreSQL adapter](postgres.md)
- [Valkey adapter](valkey.md)

## Integration recipes

- [HTTP middleware](http.md)
- [JSON-RPC middleware](json-rpc.md)
- [Queue consumers](queue.md)
- [Webhook deliveries](webhooks.md)
- [Commands and imports](commands-and-imports.md)
- [Transactions and outbox records](outbox.md)

## Project policy

- [Operations, retention, and capacity](operations.md)
- [Logging and telemetry](observability.md)
- [Troubleshooting](troubleshooting.md)
- [Migrations and compatibility](migrations-and-compatibility.md)
- [Threat model](threat-model.md)
- [Assurance review](assurance.md)
- [Resource budgets](resource-budgets.md)
- [Benchmark baselines](benchmark-baselines.md)
- [Frequently asked questions](faq.md)
- [Security policy](../SECURITY.md)
- [Contribution guide](../CONTRIBUTING.md)
- [Changelog](../CHANGELOG.md)

The published v1 API remains at its original module path. The v2 API in this
source uses `/v2` and becomes an installable release at the `v2.0.0` tag;
untagged `main` is not a supported installation. Read the state machine and
crash semantics before adopting an adapter or integration; a quickstart alone
is not a correctness specification.

## Complete documentation map

- [Durable idempotency assurance](assurance.md) — `docs/assurance.md`
- [Benchmark baselines](benchmark-baselines.md) — `docs/benchmark-baselines.md`
- [Commands and imports](commands-and-imports.md) — `docs/commands-and-imports.md`
- [Idempotency and related mechanisms](concepts.md) — `docs/concepts.md`
- [Crash and recovery semantics](crash-semantics.md) — `docs/crash-semantics.md`
- [Frequently asked questions](faq.md) — `docs/faq.md`
- [Fingerprint policy](fingerprints.md) — `docs/fingerprints.md`
- [HTTP middleware](http.md) — `docs/http.md`
- [JSON-RPC middleware](json-rpc.md) — `docs/json-rpc.md`
- [Migrations and compatibility](migrations-and-compatibility.md) — `docs/migrations-and-compatibility.md`
- [Logging and telemetry](observability.md) — `docs/observability.md`
- [Operations, retention, and capacity](operations.md) — `docs/operations.md`
- [Transactions and outbox records](outbox.md) — `docs/outbox.md`
- [Package map](package-map.md) — `docs/package-map.md`
- [PostgreSQL adapter](postgres.md) — `docs/postgres.md`
- [Queue consumer middleware](queue.md) — `docs/queue.md`
- [Five-minute quickstart](quickstart.md) — `docs/quickstart.md`
- [Resource budgets](resource-budgets.md) — `docs/resource-budgets.md`
- [Source documentation audit](source-documentation-audit.md) — `docs/source-documentation-audit.md`
- [Idempotency specification decisions](specification-decisions.md) — `docs/specification-decisions.md`
- [State machine](state-machine.md) — `docs/state-machine.md`
- [Threat model](threat-model.md) — `docs/threat-model.md`
- [Troubleshooting](troubleshooting.md) — `docs/troubleshooting.md`
- [Valkey adapter](valkey.md) — `docs/valkey.md`
- [Webhook delivery deduplication](webhooks.md) — `docs/webhooks.md`
