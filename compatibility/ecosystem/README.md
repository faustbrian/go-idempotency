# Ecosystem compatibility harness

This internal, non-releasable module compiles and exercises the published
contracts between `idempotency`, `log`, `migrations`, `transactional-outbox`,
`queue`, `telemetry`, and `webhook`. It proves independently versioned modules
can be selected together without turning them into one runtime or release unit.

The harness owns no production API. Its tests keep storage, transactions,
workers, logging, and telemetry resources caller-owned and preserve the
idempotency contract's explicit fencing and unknown-outcome boundaries.

Run the harness through the repository contract with `make check`. Shared
construction, ownership, lifecycle, and composition expectations are in the
versioned [Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.3.0/docs/ecosystem/README.md).
