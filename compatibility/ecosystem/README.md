# Ecosystem compatibility harness

This internal, non-releasable module compiles and exercises the published
contracts between `idempotency`, `log`, `migrations`, `postgres`,
`transactional-outbox`, `queue`, and `telemetry`. It proves independently versioned modules
can be selected together without turning them into one runtime or release unit.

The harness owns no production API. Its tests keep storage, transactions,
workers, logging, and telemetry resources caller-owned and preserve the
idempotency contract's explicit fencing and unknown-outcome boundaries.

## PostgreSQL transactional-service recipe

The compile-checked recipe in `contracts_test.go` is the integration boundary
for a PostgreSQL service. `go-postgres` owns pool construction and bounded
transaction cleanup; the published v4 `go-migrations` applies each module's
migrations before serving traffic; the published v3 `go-idempotency/postgres`
owns idempotency records; and the published Outbox v2
`go-transactional-outbox/v2/postgres` owns outbox
persistence. A request acquires
an idempotency lease, performs the business write, calls
`idempotencyoutbox.InsertAndComplete` with the same `pgx.Tx`, and commits once.
The application owns the transaction and must roll back on every error. A
separate outbox relay publishes committed envelopes at least once; no remote
broker or HTTP call is part of the PostgreSQL transaction.

The harness selects the published v3 Idempotency, v4 Migrations, v3
PostgreSQL, v2 Queue, v4 Telemetry and Outbox v2.1.0 modules. It asserts the
Outbox envelope writer contract against the idempotency outbox adapter and
pins the published migration checksum. The webhook replay scenario was
removed: the published Webhook v3 adapter still binds Idempotency v1, so it
cannot share this graph until Webhook adopts Idempotency v3.
This internal harness change does not release a new Idempotency module.

From the repository root, `make check` runs bounded checks across both catalog
modules, retaining their declared PostgreSQL and Valkey service prerequisites;
`make check-full` selects original full qualification. To exercise just the
public interface harness directly, run `GOWORK=off go test ./...` from this
directory. It remains non-releasable.
Shared
construction, ownership, lifecycle, and composition expectations are in the
versioned [Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.3.0/docs/ecosystem/README.md).
