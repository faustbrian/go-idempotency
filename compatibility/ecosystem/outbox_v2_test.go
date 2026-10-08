package ecosystem_test

import (
	"testing"

	idempotencyoutbox "github.com/faustbrian/go-idempotency/adapters/outbox"
	outbox "github.com/faustbrian/go-transactional-outbox/v2"
	outboxpostgres "github.com/faustbrian/go-transactional-outbox/v2/postgres"
)

// The generic idempotency adapter must accept the actual released Outbox v2
// envelope writer, not a structurally similar writer for another major.
func TestPublishedOutboxV2WriterContract(t *testing.T) {
	var _ idempotencyoutbox.Writer[outbox.Envelope] = (*outboxpostgres.Writer)(nil)
}
