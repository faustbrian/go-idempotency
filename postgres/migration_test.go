package postgres

import (
	"strings"
	"testing"
)

func TestSchemaMigrationDefinesDurableRecordAndCleanupIndex(t *testing.T) {
	migration := SchemaMigration()
	if migration.Version != 1 || migration.Name != "create_idempotency_records" {
		t.Fatalf("SchemaMigration() identity = %#v", migration)
	}
	for _, required := range []string{
		"CREATE TABLE idempotency_records",
		"record_key bytea PRIMARY KEY",
		"record jsonb NOT NULL",
		"purge_at timestamptz NOT NULL",
		"CREATE INDEX idempotency_records_purge_at_idx",
	} {
		if !strings.Contains(migration.Up, required) {
			t.Fatalf("Up migration missing %q", required)
		}
	}
	if !strings.Contains(migration.Down, "DROP TABLE idempotency_records") {
		t.Fatalf("Down migration = %q", migration.Down)
	}
}
