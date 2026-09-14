// Package idempotencyoutbox is the legacy transactional outbox adapter.
//
// Deprecated: use github.com/faustbrian/go-idempotency/adapters/outbox. This
// package remains supported for the longer of 180 days after successor
// availability and two subsequently published stable root-module minor
// releases.
package idempotencyoutbox

import (
	"context"

	"github.com/faustbrian/go-idempotency"
	canonical "github.com/faustbrian/go-idempotency/adapters/outbox"
	"github.com/jackc/pgx/v5"
)

// Writer inserts an envelope through a caller-owned transaction.
type Writer[E any] interface {
	Insert(context.Context, pgx.Tx, E) error
}

// Completer conditionally persists completion in a caller-owned transaction.
type Completer interface {
	CompleteTx(context.Context, pgx.Tx, idempotency.CompleteRequest) (idempotency.Record, error)
}

// InsertAndComplete inserts an envelope and conditionally completes its record.
func InsertAndComplete[E any](
	ctx context.Context,
	tx pgx.Tx,
	writer Writer[E],
	envelope E,
	completer Completer,
	request idempotency.CompleteRequest,
) (idempotency.Record, error) {
	return canonical.InsertAndComplete(ctx, tx, writer, envelope, completer, request)
}
