package canonical_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/faustbrian/go-idempotency/canonical"
)

func TestJSONMatchesPinnedRFC8785Fixtures(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("testdata/rfc8785.json")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var fixture struct {
		Cases []struct {
			Name     string `json:"name"`
			Input    string `json:"input"`
			Expected string `json:"expected"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	for _, test := range fixture.Cases {
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()
			actual, err := canonical.JSON([]byte(test.Input), canonical.Limits{
				MaxInputBytes: 4096, MaxOutputBytes: 4096, MaxDepth: 32,
			})
			if err != nil {
				t.Fatalf("JSON() error = %v", err)
			}
			if string(actual) != test.Expected {
				t.Fatalf("JSON() = %q, want %q", actual, test.Expected)
			}
		})
	}
}
