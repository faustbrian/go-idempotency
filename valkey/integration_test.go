package valkey

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-idempotency"
	"github.com/faustbrian/go-idempotency/idempotencytest"
	valkeygo "github.com/valkey-io/valkey-go"
)

func TestValkey9Conformance(t *testing.T) {
	client := integrationClient(t)
	runValkeyConformance(t, client, "idempotency-conformance")
}

func TestValkey9ClusterConformance(t *testing.T) {
	addresses := os.Getenv("VALKEY_CLUSTER_ADDRS")
	if addresses == "" {
		t.Skip("VALKEY_CLUSTER_ADDRS is required for Valkey 9 cluster tests")
	}
	client, err := valkeygo.NewClient(valkeygo.ClientOption{
		InitAddress: strings.Split(addresses, ","),
	})
	if err != nil {
		t.Fatalf("valkey.NewClient() error = %v", err)
	}
	t.Cleanup(client.Close)
	runValkeyConformance(t, client, "idempotency-cluster-conformance")
}

func runValkeyConformance(t *testing.T, client valkeygo.Client, prefix string) {
	t.Helper()
	idempotencytest.RunStoreConformance(t, func(t testing.TB) idempotencytest.StoreFixture {
		t.Helper()
		key, err := idempotency.NewKey("conformance", "tenant", "operation", "caller", t.Name())
		if err != nil {
			t.Fatalf("NewKey() error = %v", err)
		}
		fingerprint, err := idempotency.NewFingerprint("v1", []byte("canonical request"))
		if err != nil {
			t.Fatalf("NewFingerprint() error = %v", err)
		}
		tokens := idempotencytest.NewTokenSource("valkey-owner")
		store, err := Open(context.Background(), client, Options{
			Prefix: prefix, Retention: time.Hour, OwnerTokens: tokens.Next,
		})
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}
		storageKey := recordKey(prefix, key)
		t.Cleanup(func() {
			if err := client.Do(context.Background(), client.B().Del().Key(storageKey).Build()).Error(); err != nil {
				t.Errorf("DEL cleanup error = %v", err)
			}
		})

		return idempotencytest.StoreFixture{
			Store:       store,
			Key:         key,
			Fingerprint: fingerprint,
			Advance: func(duration time.Duration) {
				if err := client.Do(
					context.Background(),
					client.B().Hincrby().Key(storageKey).Field(fieldLeaseExpiresAt).
						Increment(-duration.Milliseconds()).Build(),
				).Error(); err != nil {
					t.Fatalf("HINCRBY lease error = %v", err)
				}
			},
		}
	})
}

func TestValkeyTTLsAndBinaryReplay(t *testing.T) {
	client := integrationClient(t)
	key, err := idempotency.NewKey("integration", "tenant", "binary", "caller", t.Name())
	if err != nil {
		t.Fatalf("NewKey() error = %v", err)
	}
	fingerprint, err := idempotency.NewFingerprint("v1", []byte("binary request"))
	if err != nil {
		t.Fatalf("NewFingerprint() error = %v", err)
	}
	tokens := idempotencytest.NewTokenSource("binary-owner")
	store, err := Open(context.Background(), client, Options{
		Prefix: "idempotency-integration", Retention: 2 * time.Minute, OwnerTokens: tokens.Next,
	})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	storageKey := recordKey("idempotency-integration", key)
	t.Cleanup(func() {
		_ = client.Do(context.Background(), client.B().Del().Key(storageKey).Build()).Error()
	})

	acquired, err := store.Acquire(context.Background(), idempotency.AcquireRequest{
		Key: key, Fingerprint: fingerprint, Lease: time.Minute,
	})
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	assertPTTL(t, client, storageKey, 3*time.Minute)
	binaryResult := []byte{0xff, 0x00, 0x01, 'x'}
	_, err = store.Complete(context.Background(), idempotency.CompleteRequest{
		Ownership: acquired.Record.Ownership(),
		Result:    binaryResult,
		Metadata:  map[string]string{"content-type": "application/octet-stream"},
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	assertPTTL(t, client, storageKey, 2*time.Minute)
	replay, err := store.Acquire(context.Background(), idempotency.AcquireRequest{
		Key: key, Fingerprint: fingerprint, Lease: time.Minute,
	})
	if err != nil {
		t.Fatalf("Acquire() replay error = %v", err)
	}
	if replay.Outcome != idempotency.OutcomeReplayed ||
		string(replay.Record.Result) != string(binaryResult) {
		t.Fatalf("replay = %#v", replay)
	}
}

func TestValkeyClientLossFailsClosed(t *testing.T) {
	address := integrationAddress(t)
	client, err := valkeygo.NewClient(valkeygo.ClientOption{InitAddress: []string{address}})
	if err != nil {
		t.Fatalf("valkey.NewClient() error = %v", err)
	}
	tokens := idempotencytest.NewTokenSource("closed-owner")
	store, err := Open(context.Background(), client, Options{
		Prefix: "idempotency-integration", Retention: time.Minute, OwnerTokens: tokens.Next,
	})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	service, err := idempotency.NewService(store)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	client.Close()
	key, err := idempotency.NewKey("integration", "tenant", "closed", "caller", t.Name())
	if err != nil {
		t.Fatalf("NewKey() error = %v", err)
	}
	fingerprint, err := idempotency.NewFingerprint("v1", []byte("closed request"))
	if err != nil {
		t.Fatalf("NewFingerprint() error = %v", err)
	}

	result, err := service.Begin(context.Background(), idempotency.BeginRequest{
		Acquire: idempotency.AcquireRequest{Key: key, Fingerprint: fingerprint, Lease: time.Minute},
	})
	if err == nil || result.Execute || result.Outcome != idempotency.OutcomeUnavailable {
		t.Fatalf("Begin() = %#v, %v", result, err)
	}
}

func TestValkeyOpenRejectsEvictingServer(t *testing.T) {
	client := integrationClient(t)
	setMaxmemoryPolicy(t, client, "allkeys-lru")
	t.Cleanup(func() {
		setMaxmemoryPolicy(t, client, "noeviction")
	})

	_, err := Open(context.Background(), client, Options{
		Prefix: "idempotency-integration", Retention: time.Minute,
		OwnerTokens: func() (string, error) { return "unused-owner", nil },
	})
	assertStoreReason(t, err, idempotency.ReasonUnsafeBackend)
}

func integrationClient(t testing.TB) valkeygo.Client {
	t.Helper()
	address := integrationAddress(t)
	client, err := valkeygo.NewClient(valkeygo.ClientOption{InitAddress: []string{address}})
	if err != nil {
		t.Fatalf("valkey.NewClient() error = %v", err)
	}
	t.Cleanup(client.Close)
	info, err := client.Do(context.Background(), client.B().Info().Section("server").Build()).ToString()
	if err != nil {
		t.Fatalf("INFO server error = %v", err)
	}
	if !strings.Contains(info, "valkey_version:9.") {
		t.Fatalf("integration server is not Valkey 9: %s", info)
	}
	return client
}

func integrationAddress(t testing.TB) string {
	t.Helper()
	address := os.Getenv("VALKEY_ADDR")
	if address == "" {
		t.Skip("VALKEY_ADDR is required for Valkey 9 integration tests")
	}
	return address
}

func assertPTTL(t *testing.T, client valkeygo.Client, key string, want time.Duration) {
	t.Helper()
	ttl, err := client.Do(context.Background(), client.B().Pttl().Key(key).Build()).AsInt64()
	if err != nil {
		t.Fatalf("PTTL error = %v", err)
	}
	minimum := want.Milliseconds() - 2_000
	if ttl < minimum || ttl > want.Milliseconds() {
		t.Fatalf("PTTL = %dms, want between %dms and %dms", ttl, minimum, want.Milliseconds())
	}
}

func setMaxmemoryPolicy(t *testing.T, client valkeygo.Client, policy string) {
	t.Helper()
	if err := client.Do(
		context.Background(),
		client.B().ConfigSet().ParameterValue().ParameterValue(
			"maxmemory-policy", policy,
		).Build(),
	).Error(); err != nil {
		t.Fatalf("CONFIG SET maxmemory-policy error = %v", err)
	}
}
