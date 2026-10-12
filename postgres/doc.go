// Package postgres implements durable idempotency storage on PostgreSQL using
// pgx, transaction-scoped advisory locks, row locks, server time, and bounded
// retention cleanup.
// Use New after applying SchemaMigration/GoMigration yourself. Pool/migrations
// and retry/deadline policy remain caller-owned. Purge ends retained identity
// and fencing history; no pool closure or cleanup scheduler is supplied.
package postgres
