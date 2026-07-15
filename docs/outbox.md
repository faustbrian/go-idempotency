# Transactions and outbox records

`postgres.Store.CompleteTx` joins idempotency completion to an
application-owned pgx transaction. Use it when the business effect or an outbox
row lives in the same PostgreSQL database.

```go
begin, err := service.Begin(ctx, idempotency.BeginRequest{
	Acquire: idempotency.AcquireRequest{
		Key: key,
		Fingerprint: fingerprint,
		Lease: 30 * time.Second,
	},
})
if err != nil {
	return err
}
if !begin.Execute {
	return handleExistingOutcome(begin)
}

tx, err := pool.Begin(ctx)
if err != nil {
	return err
}
defer tx.Rollback(context.WithoutCancel(ctx))

if err := updateBusinessRow(ctx, tx, begin.Record.FencingToken); err != nil {
	return err
}
if err := insertOutboxRecord(ctx, tx, event); err != nil {
	return err
}
if _, err := store.CompleteTx(ctx, tx, idempotency.CompleteRequest{
	Ownership: begin.Record.Ownership(),
	Result: encodedResponse,
	Metadata: map[string]string{"content-type": "application/json"},
}); err != nil {
	return err
}
if err := tx.Commit(ctx); err != nil {
	return err
}
```

The transaction takes the same advisory and row locks as ordinary completion,
rechecks ownership, fencing, state, and lease, and leaves commit or rollback to
the caller. A rollback removes both the business/outbox effect and completion.
A commit makes both visible together.

Do not call `CompleteTx` inside HTTP, JSON-RPC, queue, or command wrappers that
already complete automatically after their handler returns. Use the direct
service flow above, or a future integration mode that explicitly delegates
transaction completion.

The `go-outbox` repository currently contains planning documents but no Go
module or public record API, so a typed adapter cannot be implemented without
guessing its transaction and record contracts. Once published, its insert
operation belongs immediately before `CompleteTx` on the same `pgx.Tx`.

## Boundaries

- Acquisition commits before application work, so concurrent retries observe an
  active owner.
- The application transaction must finish before the lease expires. A stale
  transaction fails completion and must roll back its business effect.
- A commit response can be lost after PostgreSQL commits. Treat that as unknown,
  reconnect, and inspect; do not rerun based only on the transport error.
- External HTTP calls, other databases, and brokers cannot join the PostgreSQL
  transaction. Use fencing, provider idempotency keys, reconciliation, and the
  outbox publisher's own retry contract.
- Outbox delivery deduplication is separate from producer transaction atomicity.
  Consumers still need stable delivery identities or business constraints.
