package postgres

// Migration describes one reversible PostgreSQL schema change. The descriptor
// is intentionally independent until go-migrations publishes its binding API.
type Migration struct {
	Version uint
	Name    string
	Up      string
	Down    string
}

// SchemaMigration returns the initial durable-record table migration.
func SchemaMigration() Migration {
	return Migration{
		Version: 1,
		Name:    "create_idempotency_records",
		Up: `CREATE TABLE idempotency_records (
    record_key bytea PRIMARY KEY,
    record jsonb NOT NULL,
    purge_at timestamptz NOT NULL
);
CREATE INDEX idempotency_records_purge_at_idx
    ON idempotency_records (purge_at);`,
		Down: `DROP TABLE idempotency_records;`,
	}
}
