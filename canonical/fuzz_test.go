package canonical_test

import (
	"bytes"
	"testing"

	"github.com/faustbrian/go-idempotency/canonical"
)

func FuzzJSONIsIdempotent(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte(`null`),
		[]byte(`{"b":1.0,"a":[true,false]}`),
		[]byte(`{"value":-0}`),
		[]byte(`{"value":1,"value":2}`),
		[]byte(`{"music":"\ud834\udd1e"}`),
		{'"', 0xff, '"'},
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input []byte) {
		limits := canonical.Limits{
			MaxInputBytes:  4096,
			MaxOutputBytes: 4096,
			MaxDepth:       64,
		}
		canonicalValue, err := canonical.JSON(input, limits)
		if err != nil {
			return
		}
		second, err := canonical.JSON(canonicalValue, limits)
		if err != nil {
			t.Fatalf("canonical JSON was rejected: %v", err)
		}
		if !bytes.Equal(canonicalValue, second) {
			t.Fatalf("canonicalization was not idempotent: %q != %q", canonicalValue, second)
		}
	})
}
