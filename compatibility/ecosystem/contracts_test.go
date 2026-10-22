package ecosystem_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	idempotency "github.com/faustbrian/go-idempotency/v3"
	"github.com/faustbrian/go-idempotency/v3/adapters/otel"
	"github.com/faustbrian/go-idempotency/v3/adapters/outbox"
	"github.com/faustbrian/go-idempotency/v3/adapters/queue"
	"github.com/faustbrian/go-idempotency/v3/adapters/slog"
	idempotencypostgres "github.com/faustbrian/go-idempotency/v3/postgres"
	log "github.com/faustbrian/go-log/v2"
	migrations "github.com/faustbrian/go-migrations/v4"
	postgres "github.com/faustbrian/go-postgres/v3"
	queuecore "github.com/faustbrian/go-queue/v2/core"
	telemetry "github.com/faustbrian/go-telemetry/v4"
	"github.com/faustbrian/go-transactional-outbox/v2"
	outboxpostgres "github.com/faustbrian/go-transactional-outbox/v2/postgres"
	"github.com/jackc/pgx/v5"
)

func TestPublishedEcosystemContractsCompile(t *testing.T) {
	var _ idempotencyoutbox.Completer = (*idempotencypostgres.Store)(nil)
	var _ idempotencyqueue.Message = (queuecore.TaskMessage)(nil)
	var _ = idempotencyqueue.Wrap[queuecore.TaskMessage]

	logger, err := log.New(slog.NewTextHandler(io.Discard, nil))
	if err != nil {
		t.Fatalf("log.New() error = %v", err)
	}
	if _, err := idempotencylog.New(logger); err != nil {
		t.Fatalf("idempotencylog.New() error = %v", err)
	}

	_ = bindTelemetry
	_ = runTransactionalService

	// The published v3 migration must keep the ledger checksum recorded by
	// earlier majors so aligned ledgers validate without replay.
	migration, err := bindMigration()
	if err != nil {
		t.Fatalf("bindMigration() error = %v", err)
	}
	if got := migration.Checksum().String(); got != "sha256:fcdb32f4aa05f4c08b7e49e2c2acccff81eadb1320c0063ffc907e34c7d2040f" {
		t.Fatalf("migration checksum = %s", got)
	}
}

// runTransactionalService is a compile-checked composition recipe. Runtime
// tests for PostgreSQL remain in the owning modules and use their ephemeral
// database gates; this harness deliberately performs no external I/O.
func runTransactionalService(
	ctx context.Context,
	database *postgres.Pool,
	migration migrations.Migration,
	writer *outboxpostgres.Writer,
	completer idempotencyoutbox.Completer,
	envelope outbox.Envelope,
	completion idempotency.CompleteRequest,
) error {
	// Apply migration before serving requests; the caller owns the migrator.
	_ = migration
	return postgres.RunTransaction(ctx, database.Raw(), postgres.TransactionOptions{}, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "UPDATE application_records SET updated_at = now() WHERE id = $1", completion.Ownership.Key.Value()); err != nil {
			return err
		}
		_, err := idempotencyoutbox.InsertAndComplete(ctx, tx, writer, envelope, completer, completion)
		return err
	})
}

func bindMigration() (migrations.Migration, error) {
	return idempotencypostgres.GoMigration()
}

func bindTelemetry(runtime *telemetry.Runtime) (*idempotencytelemetry.Observer, error) {
	return idempotencytelemetry.New(runtime.MeterProvider())
}
