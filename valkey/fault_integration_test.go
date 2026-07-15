package valkey

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/faustbrian/go-idempotency"
	"github.com/faustbrian/go-idempotency/idempotencytest"
	valkeygo "github.com/valkey-io/valkey-go"
)

func TestValkeyUnknownResultsCanBeInspectedAfterReconnect(t *testing.T) {
	address := integrationAddress(t)
	dropper := &responseDropper{}
	client, err := valkeygo.NewClient(valkeygo.ClientOption{
		InitAddress:       []string{address},
		ForceSingleClient: true,
		PipelineMultiplex: -1,
		DisableRetry:      true,
		DialCtxFn:         dropper.dial,
	})
	if err != nil {
		t.Fatalf("valkey.NewClient() error = %v", err)
	}
	t.Cleanup(client.Close)
	store, err := Open(context.Background(), client, Options{
		Prefix: "idempotency-fault", Retention: time.Minute,
		OwnerTokens: idempotencytest.NewTokenSource("fault-owner").Next,
	})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	key, err := idempotency.NewKey("fault", "tenant", "acquire", "caller", t.Name())
	if err != nil {
		t.Fatalf("NewKey() error = %v", err)
	}
	fingerprint, err := idempotency.NewFingerprint("v1", []byte("unknown result"))
	if err != nil {
		t.Fatalf("NewFingerprint() error = %v", err)
	}
	storageKey := recordKey("idempotency-fault", key)
	t.Cleanup(func() {
		_ = client.Do(context.Background(), client.B().Del().Key(storageKey).Build()).Error()
	})

	dropper.DropNextResponse(t)
	_, acquireErr := store.Acquire(context.Background(), idempotency.AcquireRequest{
		Key: key, Fingerprint: fingerprint, Lease: time.Minute,
	})
	if acquireErr == nil {
		t.Fatal("Acquire() error = nil after response loss")
	}
	dropper.RequireDropped(t, "Acquire")

	record, err := store.Inspect(context.Background(), key)
	if err != nil {
		t.Fatalf(
			"Inspect() after reconnect error = %v; Acquire() transport error = %v",
			err, acquireErr,
		)
	}
	if record.State != idempotency.StateAcquired || record.Attempt != 1 {
		t.Fatalf("Inspect() after reconnect = %#v", record)
	}
	if record.Fingerprint != fingerprint {
		t.Fatalf("Inspect() fingerprint = %#v, want %#v", record.Fingerprint, fingerprint)
	}

	dropper.DropNextResponse(t)
	_, completeErr := store.Complete(context.Background(), idempotency.CompleteRequest{
		Ownership: record.Ownership(), Result: []byte("durable result"),
	})
	if completeErr == nil {
		t.Fatal("Complete() error = nil after response loss")
	}
	dropper.RequireDropped(t, "Complete")

	record, err = store.Inspect(context.Background(), key)
	if err != nil {
		t.Fatalf(
			"Inspect() completed record after reconnect error = %v; "+
				"Complete() transport error = %v",
			err, completeErr,
		)
	}
	if record.State != idempotency.StateCompleted || string(record.Result) != "durable result" {
		t.Fatalf("Inspect() completed record after reconnect = %#v", record)
	}
}

type responseDropper struct {
	mu      sync.Mutex
	current *dropResponseConn
}

func (d *responseDropper) dial(
	ctx context.Context,
	address string,
	dialer *net.Dialer,
	_ *tls.Config,
) (net.Conn, error) {
	connection, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}
	wrapper := &dropResponseConn{Conn: connection}
	d.mu.Lock()
	d.current = wrapper
	d.mu.Unlock()
	return wrapper, nil
}

func (d *responseDropper) DropNextResponse(t *testing.T) {
	t.Helper()
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.current == nil {
		t.Fatal("response dropper has no connection")
	}
	d.current.dropped.Store(false)
	d.current.armed.Store(true)
}

func (d *responseDropper) RequireDropped(t *testing.T, operation string) {
	t.Helper()
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.current == nil || !d.current.dropped.Load() {
		t.Fatalf("%s failed before the injected response-loss boundary", operation)
	}
}

type dropResponseConn struct {
	net.Conn
	armed   atomic.Bool
	dropped atomic.Bool
}

func (c *dropResponseConn) Read(buffer []byte) (int, error) {
	read, err := c.Conn.Read(buffer)
	if read > 0 && c.armed.CompareAndSwap(true, false) {
		c.dropped.Store(true)
		_ = c.Conn.Close()
		return 0, io.ErrUnexpectedEOF
	}
	return read, err
}
