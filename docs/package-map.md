# Package map

[Documentation index](README.md) · [Repository entry](../README.md)

This map follows the current source modules, not a synchronized release.
Use the repository README for published installation versions; an unpublished
source major is not installable merely because its package appears here.
Internal packages and non-releasable fixtures are intentionally excluded.

## github.com/faustbrian/go-idempotency/v3

Provide durable ownership, fencing, and bounded result replay for retried HTTP, JSON-RPC, webhook, queue, import, and command workloads.

| Import path | Responsibility |
| --- | --- |
| [`github.com/faustbrian/go-idempotency/v3`](../doc.go) | Use for durable ownership, fencing, typed outcomes, bounded result replay, and the backend-neutral service contract. |
| [`github.com/faustbrian/go-idempotency/v3/adapters/command`](../adapters/command/runner.go) | Use to apply durable ownership and replay to command and import handlers. |
| [`github.com/faustbrian/go-idempotency/v3/adapters/http`](../adapters/http/doc.go) | Use to apply Idempotency-Key policy and bounded response replay to HTTP handlers. |
| [`github.com/faustbrian/go-idempotency/v3/adapters/jsonrpc`](../adapters/jsonrpc/middleware.go) | Use to apply method-aware request ownership and response replay to JSON-RPC handlers. |
| [`github.com/faustbrian/go-idempotency/v3/adapters/otel`](../adapters/otel/observer.go) | Use for bounded OpenTelemetry transition metrics. |
| [`github.com/faustbrian/go-idempotency/v3/adapters/outbox`](../adapters/outbox/coordinator.go) | Use to stage an outbox envelope and idempotency completion in one caller-owned transaction. |
| [`github.com/faustbrian/go-idempotency/v3/adapters/queue`](../adapters/queue/middleware.go) | Use to apply durable ownership and redelivery deduplication to queue messages. |
| [`github.com/faustbrian/go-idempotency/v3/adapters/slog`](../adapters/slog/observer.go) | Use for secret-safe structured transition logging through log/slog. |
| [`github.com/faustbrian/go-idempotency/v3/adapters/webhook`](../adapters/webhook/processor.go) | Use for provider-delivery deduplication after signature verification. |
| [`github.com/faustbrian/go-idempotency/v3/canonical`](../canonical/canonical.go) | Use for bounded canonical JSON and byte fingerprint construction. |
| [`github.com/faustbrian/go-idempotency/v3/idempotencytest`](../idempotencytest/conformance.go) | Use in tests for deterministic clocks, tokens, fixtures, and backend conformance. |
| [`github.com/faustbrian/go-idempotency/v3/memory`](../memory/store.go) | Use for deterministic in-process storage where restart loss is acceptable. |
| [`github.com/faustbrian/go-idempotency/v3/postgres`](../postgres/doc.go) | Use for durable PostgreSQL storage and caller-owned transactional completion. |
| [`github.com/faustbrian/go-idempotency/v3/valkey`](../valkey/codec.go) | Use for durable low-latency ownership in a dedicated Valkey noeviction deployment. |

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
