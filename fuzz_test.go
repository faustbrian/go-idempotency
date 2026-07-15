package idempotency_test

import (
	"testing"

	"github.com/faustbrian/go-idempotency"
)

func FuzzKeyAndFingerprintValidation(f *testing.F) {
	f.Add(
		"http", "tenant", "POST /widgets", "caller", "request-1",
		"json-v1", []byte(`{"amount":42}`), make([]byte, 32),
	)
	f.Add("", "", "", "", "", "", []byte{0xff}, []byte{0x00})

	f.Fuzz(func(
		t *testing.T,
		namespace string,
		tenant string,
		operation string,
		caller string,
		value string,
		version string,
		canonical []byte,
		persisted []byte,
	) {
		parts := []string{namespace, tenant, operation, caller, value}
		for _, part := range parts {
			if len(part) > idempotency.MaxKeyPartBytes+1 {
				return
			}
		}
		if len(canonical) > 4096 || len(persisted) > 64 {
			return
		}

		key, err := idempotency.NewKey(namespace, tenant, operation, caller, value)
		validKey := true
		for _, part := range parts {
			validKey = validKey && part != "" && len(part) <= idempotency.MaxKeyPartBytes
		}
		if validKey != (err == nil) {
			t.Fatalf("NewKey() valid = %t, error = %v", validKey, err)
		}
		if err == nil && (key.Namespace() != namespace || key.Tenant() != tenant ||
			key.Operation() != operation || key.Caller() != caller || key.Value() != value) {
			t.Fatalf("NewKey() did not preserve all identity parts: %#v", key)
		}

		fingerprint, err := idempotency.NewFingerprint(version, canonical)
		if (version != "") != (err == nil) {
			t.Fatalf("NewFingerprint() version = %q, error = %v", version, err)
		}
		if err == nil && (fingerprint.Version() != version || len(fingerprint.Sum()) != 32) {
			t.Fatalf("NewFingerprint() = %#v", fingerprint)
		}

		reconstructed, err := idempotency.NewFingerprintFromSum(version, persisted)
		validPersisted := version != "" && len(persisted) == 32
		if validPersisted != (err == nil) {
			t.Fatalf("NewFingerprintFromSum() valid = %t, error = %v", validPersisted, err)
		}
		if err == nil && (reconstructed.Version() != version ||
			string(reconstructed.Sum()) != string(persisted)) {
			t.Fatalf("NewFingerprintFromSum() = %#v", reconstructed)
		}
	})
}
