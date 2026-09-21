package valkey

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-idempotency"
	"github.com/faustbrian/go-idempotency/idempotencytest"
	valkeygo "github.com/valkey-io/valkey-go"
)

func TestValkeyRejectsUnboundedOrMalformedPersistedFieldsBeforeEveryOperation(t *testing.T) {
	fieldLimits := map[string]int{
		fieldSchema: 1, fieldNamespace: idempotency.MaxKeyPartBytes,
		fieldTenant: idempotency.MaxKeyPartBytes, fieldOperation: idempotency.MaxKeyPartBytes,
		fieldCaller: idempotency.MaxKeyPartBytes, fieldKeyValue: idempotency.MaxKeyPartBytes,
		fieldFingerprintVersion: idempotency.MaxFingerprintVersionBytes,
		fieldFingerprintSum:     64, fieldState: 9, fieldOwnerToken: idempotency.MaxOwnerTokenBytes,
		fieldFencingToken: 20, fieldLeaseExpiresAt: 19, fieldHeartbeatAt: 19,
		fieldAttempt: 20, fieldCreatedAt: 19, fieldUpdatedAt: 19,
		fieldCompletedAt: 19, fieldFailedAt: 19, fieldAbandonedAt: 19,
		fieldExpiredAt: 19, fieldResult: idempotency.MaxResultBytes,
		fieldMetadata: maxEncodedMetadataBytes,
	}
	semanticCorruptions := map[string]string{
		fieldSchema: "2", fieldNamespace: "different", fieldTenant: "different",
		fieldOperation: "different", fieldCaller: "different", fieldKeyValue: "different",
		fieldFingerprintVersion: "", fieldFingerprintSum: "00", fieldState: "future",
		fieldOwnerToken: "", fieldFencingToken: "0", fieldLeaseExpiresAt: "later",
		fieldHeartbeatAt: "-1", fieldAttempt: "18446744073709551616",
		fieldCreatedAt: "later", fieldUpdatedAt: "later", fieldCompletedAt: "later",
		fieldFailedAt: "later", fieldAbandonedAt: "later", fieldExpiredAt: "later",
		fieldMetadata: `{"key":1}`,
	}

	for _, protocol := range []struct {
		name  string
		resp2 bool
	}{{name: "resp3"}, {name: "resp2", resp2: true}} {
		t.Run(protocol.name, func(t *testing.T) {
			client := validationClient(t, protocol.resp2)
			for field, limit := range fieldLimits {
				field, limit := field, limit
				for _, operation := range validationOperations() {
					operation := operation
					t.Run("oversized/"+field+"/"+string(operation), func(t *testing.T) {
						fixture := newValidationFixture(t, client, field+string(operation))
						setHashField(t, client, fixture.storageKey, field, strings.Repeat("x", limit+1))
						before := snapshotValue(t, client, fixture.storageKey)
						err := fixture.execute(operation)
						assertStoreReason(t, err, idempotency.ReasonLimitExceeded)
						assertSnapshot(t, client, fixture.storageKey, before)
					})
				}
			}
			for field, value := range semanticCorruptions {
				field, value := field, value
				for _, operation := range validationOperations() {
					operation := operation
					t.Run("malformed/"+field+"/"+string(operation), func(t *testing.T) {
						fixture := newValidationFixture(t, client, field+string(operation))
						setHashField(t, client, fixture.storageKey, field, value)
						before := snapshotValue(t, client, fixture.storageKey)
						err := fixture.execute(operation)
						assertStoreReason(t, err, idempotency.ReasonInvalidPayload)
						assertSnapshot(t, client, fixture.storageKey, before)
					})
				}
			}
		})
	}
}

func TestValkeyRejectsMalformedRecordShapesAndWrongTypesWithoutMutation(t *testing.T) {
	for _, protocol := range []struct {
		name  string
		resp2 bool
	}{{name: "resp3"}, {name: "resp2", resp2: true}} {
		t.Run(protocol.name, func(t *testing.T) {
			client := validationClient(t, protocol.resp2)
			for _, operation := range validationOperations() {
				operation := operation
				t.Run("missing/"+string(operation), func(t *testing.T) {
					fixture := newValidationFixture(t, client, "missing"+string(operation))
					deleteHashField(t, client, fixture.storageKey, fieldMetadata)
					before := snapshotValue(t, client, fixture.storageKey)
					assertStoreReason(t, fixture.execute(operation), idempotency.ReasonInvalidPayload)
					assertSnapshot(t, client, fixture.storageKey, before)
				})
				t.Run("substituted/"+string(operation), func(t *testing.T) {
					fixture := newValidationFixture(t, client, "substituted"+string(operation))
					deleteHashField(t, client, fixture.storageKey, fieldMetadata)
					setHashField(t, client, fixture.storageKey, "unknown", "{}")
					before := snapshotValue(t, client, fixture.storageKey)
					assertStoreReason(t, fixture.execute(operation), idempotency.ReasonInvalidPayload)
					assertSnapshot(t, client, fixture.storageKey, before)
				})
				t.Run("extra/"+string(operation), func(t *testing.T) {
					fixture := newValidationFixture(t, client, "extra"+string(operation))
					setHashField(t, client, fixture.storageKey, "unknown", "value")
					before := snapshotValue(t, client, fixture.storageKey)
					assertStoreReason(t, fixture.execute(operation), idempotency.ReasonInvalidPayload)
					assertSnapshot(t, client, fixture.storageKey, before)
				})
				t.Run("wrong-type/"+string(operation), func(t *testing.T) {
					fixture := newValidationFixture(t, client, "wrong-type"+string(operation))
					if err := client.Do(
						context.Background(), client.B().Del().Key(fixture.storageKey).Build(),
					).Error(); err != nil {
						t.Fatalf("DEL error = %v", err)
					}
					if err := client.Do(
						context.Background(),
						client.B().Set().Key(fixture.storageKey).Value("hostile").Build(),
					).Error(); err != nil {
						t.Fatalf("SET error = %v", err)
					}
					if err := client.Do(
						context.Background(), client.B().Pexpire().Key(fixture.storageKey).Milliseconds(60_000).Build(),
					).Error(); err != nil {
						t.Fatalf("PEXPIRE error = %v", err)
					}
					before := snapshotValue(t, client, fixture.storageKey)
					assertStoreReason(t, fixture.execute(operation), idempotency.ReasonInvalidPayload)
					assertSnapshot(t, client, fixture.storageKey, before)
				})
			}
		})
	}
}

func TestValkeyPreservesNullAndEmptyMetadata(t *testing.T) {
	for _, protocol := range []struct {
		name  string
		resp2 bool
	}{{name: "resp3"}, {name: "resp2", resp2: true}} {
		t.Run(protocol.name, func(t *testing.T) {
			client := validationClient(t, protocol.resp2)
			for _, terminal := range []struct {
				name     string
				metadata map[string]string
				fail     bool
			}{{name: "null"}, {name: "empty", metadata: map[string]string{}},
				{name: "failed-null", fail: true}, {name: "failed-empty", fail: true, metadata: map[string]string{}}} {
				terminal := terminal
				t.Run(terminal.name, func(t *testing.T) {
					fixture := newValidationFixture(t, client, terminal.name)
					var record idempotency.Record
					var err error
					if terminal.fail {
						record, err = fixture.store.Fail(context.Background(), idempotency.FailRequest{
							Ownership: fixture.acquired.Record.Ownership(), Metadata: terminal.metadata,
						})
					} else {
						record, err = fixture.store.Complete(context.Background(), idempotency.CompleteRequest{
							Ownership: fixture.acquired.Record.Ownership(), Metadata: terminal.metadata,
						})
					}
					if err != nil {
						t.Fatalf("terminal mutation error = %v", err)
					}
					assertMetadataRepresentation(t, record.Metadata, terminal.metadata == nil)
					inspected, err := fixture.store.Inspect(context.Background(), fixture.key)
					if err != nil {
						t.Fatalf("Inspect() error = %v", err)
					}
					assertMetadataRepresentation(t, inspected.Metadata, terminal.metadata == nil)
					replayed, err := fixture.store.Acquire(context.Background(), idempotency.AcquireRequest{
						Key: fixture.key, Fingerprint: fixture.fingerprint, Lease: time.Minute,
					})
					if err != nil {
						t.Fatalf("Acquire(replay) error = %v", err)
					}
					assertMetadataRepresentation(t, replayed.Record.Metadata, terminal.metadata == nil)
				})
			}
		})
	}
}

func TestValkeyAcceptsSemanticPersistenceBoundaries(t *testing.T) {
	for _, protocol := range []struct {
		name  string
		resp2 bool
	}{{name: "resp3"}, {name: "resp2", resp2: true}} {
		t.Run(protocol.name, func(t *testing.T) {
			client := validationClient(t, protocol.resp2)
			part := strings.Repeat("k", idempotency.MaxKeyPartBytes)
			key, err := idempotency.NewKey(part, part, part, part, part)
			if err != nil {
				t.Fatalf("NewKey() error = %v", err)
			}
			fingerprint, err := idempotency.NewFingerprint(
				strings.Repeat("v", idempotency.MaxFingerprintVersionBytes), []byte("boundary"),
			)
			if err != nil {
				t.Fatalf("NewFingerprint() error = %v", err)
			}
			owner := strings.Repeat("o", idempotency.MaxOwnerTokenBytes)
			store, err := New(client, Options{
				Prefix: "idempotency-boundary", Retention: MaxRetention,
				OwnerTokens: func() (string, error) { return owner, nil },
			})
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			storageKey := recordKey("idempotency-boundary", key)
			t.Cleanup(func() {
				_ = client.Do(context.Background(), client.B().Del().Key(storageKey).Build()).Error()
			})
			acquired, err := store.Acquire(context.Background(), idempotency.AcquireRequest{
				Key: key, Fingerprint: fingerprint, Lease: idempotency.MaxLease,
			})
			if err != nil {
				t.Fatalf("Acquire() error = %v", err)
			}
			metadata := make(map[string]string, idempotency.MaxMetadataEntries)
			for index := range idempotency.MaxMetadataEntries {
				prefix := fmt.Sprintf("%02d", index)
				metadata[prefix+strings.Repeat("k", idempotency.MaxMetadataKeyBytes-2)] =
					strings.Repeat("v", idempotency.MaxMetadataValueBytes)
			}
			completed, err := store.Complete(context.Background(), idempotency.CompleteRequest{
				Ownership: acquired.Record.Ownership(),
				Result:    make([]byte, idempotency.MaxResultBytes),
				Metadata:  metadata,
			})
			if err != nil {
				t.Fatalf("Complete() error = %v", err)
			}
			if completed.State != idempotency.StateCompleted ||
				len(completed.OwnerToken) != idempotency.MaxOwnerTokenBytes ||
				len(completed.Result) != idempotency.MaxResultBytes ||
				len(completed.Metadata) != idempotency.MaxMetadataEntries {
				t.Fatalf("Complete() boundary record = %#v", completed)
			}

			setHashField(t, client, storageKey, fieldState, string(idempotency.StateAbandoned))
			setHashField(t, client, storageKey, fieldFencingToken, "18446744073709551615")
			setHashField(t, client, storageKey, fieldAttempt, "18446744073709551615")
			for _, field := range []string{
				fieldLeaseExpiresAt, fieldHeartbeatAt, fieldCreatedAt, fieldUpdatedAt,
				fieldCompletedAt, fieldFailedAt, fieldAbandonedAt, fieldExpiredAt,
			} {
				setHashField(t, client, storageKey, field, "9223372036854775807")
			}
			inspected, err := store.Inspect(context.Background(), key)
			if err != nil {
				t.Fatalf("Inspect() semantic maxima error = %v", err)
			}
			if inspected.FencingToken != ^uint64(0) || inspected.Attempt != ^uint64(0) ||
				inspected.State != idempotency.StateAbandoned {
				t.Fatalf("Inspect() semantic maxima = %#v", inspected)
			}
		})
	}
}

func TestValkeyTakeoverAcceptsSignedCounterBoundaryAndRejectsOverflow(t *testing.T) {
	client := validationClient(t, false)
	for _, test := range []struct {
		name   string
		value  string
		reason idempotency.Reason
	}{
		{name: "maximum incrementable", value: "9223372036854775806"},
		{name: "overflow", value: "9223372036854775807", reason: idempotency.ReasonLimitExceeded},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := newValidationFixture(t, client, test.name)
			setHashField(t, client, fixture.storageKey, fieldFencingToken, test.value)
			setHashField(t, client, fixture.storageKey, fieldAttempt, test.value)
			setHashField(t, client, fixture.storageKey, fieldLeaseExpiresAt, "1")
			before := snapshotValue(t, client, fixture.storageKey)
			result, err := fixture.store.Acquire(context.Background(), idempotency.AcquireRequest{
				Key: fixture.key, Fingerprint: fixture.fingerprint, Lease: time.Minute,
			})
			if test.reason != "" {
				assertStoreReason(t, err, test.reason)
				assertSnapshot(t, client, fixture.storageKey, before)
				return
			}
			if err != nil || result.Outcome != idempotency.OutcomeStaleOwnerTakeover ||
				result.Record.FencingToken != uint64(9223372036854775807) ||
				result.Record.Attempt != uint64(9223372036854775807) {
				t.Fatalf("Acquire() counter boundary = %#v, %v", result, err)
			}
		})
	}
}

type validationFixture struct {
	store       *Store
	key         idempotency.Key
	fingerprint idempotency.Fingerprint
	acquired    idempotency.AcquireResult
	storageKey  string
}

func newValidationFixture(t *testing.T, client valkeygo.Client, suffix string) validationFixture {
	t.Helper()
	prefix := "idempotency-validation"
	key, err := idempotency.NewKey("validation", "tenant", "operation", "caller", t.Name()+suffix)
	if err != nil {
		t.Fatalf("NewKey() error = %v", err)
	}
	fingerprint, err := idempotency.NewFingerprint("v1", []byte("validation"))
	if err != nil {
		t.Fatalf("NewFingerprint() error = %v", err)
	}
	store, err := New(client, Options{
		Prefix: prefix, Retention: time.Hour,
		OwnerTokens: idempotencytest.NewTokenSource("validation-owner").Next,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	acquired, err := store.Acquire(context.Background(), idempotency.AcquireRequest{
		Key: key, Fingerprint: fingerprint, Lease: time.Minute,
	})
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	storageKey := recordKey(prefix, key)
	t.Cleanup(func() {
		_ = client.Do(context.Background(), client.B().Del().Key(storageKey).Build()).Error()
	})
	return validationFixture{
		store: store, key: key, fingerprint: fingerprint, acquired: acquired, storageKey: storageKey,
	}
}

func (f validationFixture) execute(operation operation) error {
	ctx := context.Background()
	switch operation {
	case operationAcquire:
		_, err := f.store.Acquire(ctx, idempotency.AcquireRequest{
			Key: f.key, Fingerprint: f.fingerprint, Lease: time.Minute,
		})
		return err
	case operationInspect:
		_, err := f.store.Inspect(ctx, f.key)
		return err
	case operationHeartbeat:
		_, err := f.store.Heartbeat(ctx, idempotency.HeartbeatRequest{
			Ownership: f.acquired.Record.Ownership(), Lease: time.Minute,
		})
		return err
	case operationComplete:
		_, err := f.store.Complete(ctx, idempotency.CompleteRequest{Ownership: f.acquired.Record.Ownership()})
		return err
	case operationFail:
		_, err := f.store.Fail(ctx, idempotency.FailRequest{Ownership: f.acquired.Record.Ownership()})
		return err
	case operationRelease:
		_, err := f.store.Release(ctx, f.acquired.Record.Ownership())
		return err
	case operationExpire:
		_, err := f.store.Expire(ctx, f.key)
		return err
	default:
		return &idempotency.Error{Reason: idempotency.ReasonInvalidPayload, Field: "test_operation"}
	}
}

func validationOperations() []operation {
	return []operation{
		operationAcquire, operationInspect, operationHeartbeat, operationComplete,
		operationFail, operationRelease, operationExpire,
	}
}

func validationClient(t *testing.T, resp2 bool) valkeygo.Client {
	t.Helper()
	client, err := valkeygo.NewClient(valkeygo.ClientOption{
		InitAddress: []string{integrationAddress(t)}, AlwaysRESP2: resp2,
		DisableCache: resp2,
	})
	if err != nil {
		t.Fatalf("valkey.NewClient() error = %v", err)
	}
	t.Cleanup(client.Close)
	return client
}

type valueSnapshot struct {
	dump   string
	expiry int64
}

func snapshotValue(t *testing.T, client valkeygo.Client, key string) valueSnapshot {
	t.Helper()
	dump, err := client.Do(context.Background(), client.B().Dump().Key(key).Build()).ToString()
	if err != nil {
		t.Fatalf("DUMP error = %v", err)
	}
	expiry, err := client.Do(context.Background(), client.B().Pexpiretime().Key(key).Build()).AsInt64()
	if err != nil {
		t.Fatalf("PEXPIRETIME error = %v", err)
	}
	return valueSnapshot{dump: dump, expiry: expiry}
}

func assertSnapshot(t *testing.T, client valkeygo.Client, key string, want valueSnapshot) {
	t.Helper()
	got := snapshotValue(t, client, key)
	if got != want {
		t.Fatalf("record changed: dump_equal=%t expiry=%d want=%d", got.dump == want.dump, got.expiry, want.expiry)
	}
}

func setHashField(t *testing.T, client valkeygo.Client, key, field, value string) {
	t.Helper()
	if err := client.Do(
		context.Background(), client.B().Hset().Key(key).FieldValue().FieldValue(field, value).Build(),
	).Error(); err != nil {
		t.Fatalf("HSET %s error = %v", field, err)
	}
}

func deleteHashField(t *testing.T, client valkeygo.Client, key, field string) {
	t.Helper()
	if err := client.Do(
		context.Background(), client.B().Hdel().Key(key).Field(field).Build(),
	).Error(); err != nil {
		t.Fatalf("HDEL %s error = %v", field, err)
	}
}

func assertMetadataRepresentation(t *testing.T, got map[string]string, wantNil bool) {
	t.Helper()
	if wantNil && got != nil {
		t.Fatalf("metadata = %#v, want nil", got)
	}
	if !wantNil && (got == nil || len(got) != 0) {
		t.Fatalf("metadata = %#v, want non-nil empty map", got)
	}
}
