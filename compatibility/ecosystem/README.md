# Ecosystem compatibility harness

This internal, non-releasable module compiles and exercises the published
contracts between `idempotency`, `log`, `migrations`, `transactional-outbox`,
`queue`, `telemetry`, and `webhook`. It proves independently versioned modules
can be selected together without turning them into one runtime or release unit.

The harness owns no production API. Its tests keep storage, transactions,
workers, logging, and telemetry resources caller-owned and preserve the
idempotency contract's explicit fencing and unknown-outcome boundaries.

## PostgreSQL transactional-service recipe

The compile-checked recipe in `contracts_test.go` is the integration boundary
for a PostgreSQL service. `go-postgres` owns pool construction and bounded
transaction cleanup; the published v1 `go-migrations` applies each module's
migrations before serving traffic; the published v1 `go-idempotency/postgres`
owns idempotency records; and the published Outbox v2
`go-transactional-outbox/v2/postgres` owns outbox
persistence. A request acquires
an idempotency lease, performs the business write, calls
`idempotencyoutbox.InsertAndComplete` with the same `pgx.Tx`, and commits once.
The application owns the transaction and must roll back on every error. A
separate outbox relay publishes committed envelopes at least once; no remote
broker or HTTP call is part of the PostgreSQL transaction.

The harness selects public Outbox v2.0.0 and asserts its exact envelope writer
contract against the generic v1 idempotency adapter. Webhook remains on its
published v2.0.0 line here until the separate Webhook v3 release is delivered.
This internal harness change does not release a new Idempotency module.

Run the harness through the repository contract with `make check`. Shared
construction, ownership, lifecycle, and composition expectations are in the
versioned [Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.3.0/docs/ecosystem/README.md).
