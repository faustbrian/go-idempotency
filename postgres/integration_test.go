package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/faustbrian/go-idempotency"
	"github.com/faustbrian/go-idempotency/idempotencytest"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresConformance(t *testing.T) {
	pool := integrationPool(t)
	idempotencytest.RunStoreConformance(t, func(t testing.TB) idempotencytest.StoreFixture {
		key, err := idempotency.NewKey("postgres-live", "tenant", "operation", "caller", t.Name())
		if err != nil {
			t.Fatalf("NewKey() error = %v", err)
		}
		fingerprint, err := idempotency.NewFingerprint("v1", []byte("request"))
		if err != nil {
			t.Fatalf("NewFingerprint() error = %v", err)
		}
		store, err := New(pool, Options{
			Retention:   time.Hour,
			OwnerTokens: idempotencytest.NewTokenSource("postgres-live-owner").Next,
		})
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		digest := recordDigest(key)
		t.Cleanup(func() {
			if _, err := pool.Exec(
				context.Background(), deleteRecordSQL, digest,
			); err != nil {
				t.Errorf("record cleanup error = %v", err)
			}
		})
		return idempotencytest.StoreFixture{
			Store: store, Key: key, Fingerprint: fingerprint,
			Advance: func(duration time.Duration) {
				_, err := pool.Exec(context.Background(),
					"UPDATE idempotency_records SET record = jsonb_set("+
						"record, '{lease_expires_at}', to_jsonb(to_char("+
						"(record->>'lease_expires_at')::timestamptz - "+
						"make_interval(secs => $2), "+
						"'YYYY-MM-DD\"T\"HH24:MI:SS.US\"Z\"'))) "+
						"WHERE record_key = $1",
					digest, duration.Seconds(),
				)
				if err != nil {
					t.Fatalf("lease advance error = %v", err)
				}
			},
		}
	})
}

func TestPostgresCleanupRemovesExpiredRetentionRows(t *testing.T) {
	pool := integrationPool(t)
	store, err := New(pool, Options{
		Retention:   time.Hour,
		OwnerTokens: idempotencytest.NewTokenSource("cleanup-owner").Next,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	key, fingerprint := storeIdentity(t, t.Name())
	acquired, err := store.Acquire(context.Background(), idempotency.AcquireRequest{
		Key: key, Fingerprint: fingerprint, Lease: time.Minute,
	})
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	if _, err := store.Complete(context.Background(), idempotency.CompleteRequest{
		Ownership: acquired.Record.Ownership(), Result: []byte{0xff, 0x00, 0x01},
	}); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	replayed, err := store.Acquire(context.Background(), idempotency.AcquireRequest{
		Key: key, Fingerprint: fingerprint, Lease: time.Minute,
	})
	if err != nil || replayed.Outcome != idempotency.OutcomeReplayed ||
		string(replayed.Record.Result) != string([]byte{0xff, 0x00, 0x01}) {
		t.Fatalf("Acquire() binary replay = %#v, %v", replayed, err)
	}
	if _, err := pool.Exec(context.Background(),
		"UPDATE idempotency_records SET purge_at = clock_timestamp() - interval '1 second' "+
			"WHERE record_key = $1",
		recordDigest(key),
	); err != nil {
		t.Fatalf("expire purge_at error = %v", err)
	}
	count, err := store.Cleanup(context.Background(), 10)
	if err != nil || count != 1 {
		t.Fatalf("Cleanup() = %d, %v", count, err)
	}
	if _, err := store.Inspect(context.Background(), key); err == nil {
		t.Fatal("Inspect() cleaned record error = nil")
	}
}

func TestPostgresMalformedRecordAndClosedPoolFailClosed(t *testing.T) {
	pool := integrationPool(t)
	store, err := New(pool, Options{
		Retention:   time.Hour,
		OwnerTokens: idempotencytest.NewTokenSource("failure-owner").Next,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	key, fingerprint := storeIdentity(t, t.Name())
	if _, err := pool.Exec(context.Background(), upsertRecordSQL,
		recordDigest(key), []byte(`{"schema":2}`), time.Now().Add(time.Hour),
	); err != nil {
		t.Fatalf("insert malformed record error = %v", err)
	}
	if _, err := store.Inspect(context.Background(), key); err == nil {
		t.Fatal("Inspect() malformed record error = nil")
	}
	if _, err := pool.Exec(
		context.Background(), deleteRecordSQL, recordDigest(key),
	); err != nil {
		t.Fatalf("delete malformed record error = %v", err)
	}
	service, err := idempotency.NewService(store)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	pool.Close()
	begin, err := service.Begin(context.Background(), idempotency.BeginRequest{
		Acquire: idempotency.AcquireRequest{
			Key: key, Fingerprint: fingerprint, Lease: time.Minute,
		},
	})
	if err == nil || begin.Execute || begin.Outcome != idempotency.OutcomeUnavailable {
		t.Fatalf("Begin() closed pool = %#v, %v", begin, err)
	}
}

func TestPostgresCompleteTxCommitsWithBusinessEffect(t *testing.T) {
	pool := integrationPool(t)
	if _, err := pool.Exec(context.Background(),
		"CREATE TABLE business_effects (id text PRIMARY KEY)",
	); err != nil {
		t.Fatalf("create business_effects error = %v", err)
	}
	store, err := New(pool, Options{
		Retention:   time.Hour,
		OwnerTokens: idempotencytest.NewTokenSource("transaction-owner").Next,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	key, fingerprint := storeIdentity(t, t.Name())
	acquired, err := store.Acquire(context.Background(), idempotency.AcquireRequest{
		Key: key, Fingerprint: fingerprint, Lease: time.Minute,
	})
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}

	tx, err := pool.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin() rollback transaction error = %v", err)
	}
	if _, err := tx.Exec(context.Background(),
		"INSERT INTO business_effects (id) VALUES ($1)", "rolled-back",
	); err != nil {
		t.Fatalf("insert rollback effect error = %v", err)
	}
	if _, err := store.CompleteTx(context.Background(), tx, idempotency.CompleteRequest{
		Ownership: acquired.Record.Ownership(), Result: []byte("rolled-back"),
	}); err != nil {
		t.Fatalf("CompleteTx() rollback error = %v", err)
	}
	if err := tx.Rollback(context.Background()); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}
	rolledBack, err := store.Inspect(context.Background(), key)
	if err != nil || rolledBack.State != idempotency.StateAcquired {
		t.Fatalf("Inspect() after rollback = %#v, %v", rolledBack, err)
	}

	tx, err = pool.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin() commit transaction error = %v", err)
	}
	if _, err := tx.Exec(context.Background(),
		"INSERT INTO business_effects (id) VALUES ($1)", "committed",
	); err != nil {
		t.Fatalf("insert committed effect error = %v", err)
	}
	if _, err := store.CompleteTx(context.Background(), tx, idempotency.CompleteRequest{
		Ownership: acquired.Record.Ownership(), Result: []byte("committed"),
	}); err != nil {
		t.Fatalf("CompleteTx() commit error = %v", err)
	}
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	completed, err := store.Inspect(context.Background(), key)
	if err != nil || completed.State != idempotency.StateCompleted ||
		string(completed.Result) != "committed" {
		t.Fatalf("Inspect() after commit = %#v, %v", completed, err)
	}
	var effects int
	if err := pool.QueryRow(context.Background(),
		"SELECT count(*) FROM business_effects",
	).Scan(&effects); err != nil || effects != 1 {
		t.Fatalf("business effect count = %d, %v", effects, err)
	}
}

func integrationPool(t testing.TB) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("POSTGRES_URL")
	if databaseURL == "" {
		t.Skip("POSTGRES_URL is required for PostgreSQL integration tests")
	}
	suffix := make([]byte, 8)
	if _, err := rand.Read(suffix); err != nil {
		t.Fatalf("rand.Read() error = %v", err)
	}
	schema := "idempotency_test_" + hex.EncodeToString(suffix)
	bootstrap, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("pgxpool.New() bootstrap error = %v", err)
	}
	if _, err := bootstrap.Exec(
		context.Background(), "CREATE SCHEMA "+pgx.Identifier{schema}.Sanitize(),
	); err != nil {
		bootstrap.Close()
		t.Fatalf("CREATE SCHEMA error = %v", err)
	}
	bootstrap.Close()

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("pgxpool.ParseConfig() error = %v", err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Fatalf("pgxpool.NewWithConfig() error = %v", err)
	}
	if _, err := pool.Exec(context.Background(), SchemaMigration().Up); err != nil {
		pool.Close()
		t.Fatalf("schema migration error = %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
		admin, err := pgxpool.New(context.Background(), databaseURL)
		if err != nil {
			t.Errorf("cleanup pgxpool.New() error = %v", err)
			return
		}
		defer admin.Close()
		_, err = admin.Exec(
			context.Background(), "DROP SCHEMA "+pgx.Identifier{schema}.Sanitize()+" CASCADE",
		)
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("DROP SCHEMA error = %v", err)
		}
	})
	return pool
}
