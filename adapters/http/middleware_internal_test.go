package idempotencyhttp

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/faustbrian/go-idempotency/v2"
)

func TestDecodeSnapshotRejectsOversizedRecordBeforeJSONDecode(t *testing.T) {
	_, err := decodeSnapshot(make([]byte, idempotency.MaxResultBytes+1))
	var semantic *idempotency.Error
	if !errors.As(err, &semantic) || semantic.Reason != idempotency.ReasonLimitExceeded {
		t.Fatalf("decodeSnapshot() error = %#v", err)
	}
}

func TestReplayHeadersFitRejectsAggregateBeforeAddingNextName(t *testing.T) {
	values := make([]string, 8)
	remaining := MaxReplayHeaderBytes - len("A") - 1
	for index := range values {
		size := min(remaining, MaxReplayHeaderValueBytes)
		values[index] = strings.Repeat("x", size)
		remaining -= size
	}
	header := http.Header{
		"A":         values,
		"Long-Name": {""},
	}
	if replayHeadersFit(header, []string{"A", "Long-Name"}) {
		t.Fatal("replayHeadersFit() = true")
	}
}

func TestReplayHeadersFitHonorsExactResourceBoundaries(t *testing.T) {
	exactNames := make(http.Header, MaxReplayHeaderNames)
	names := make([]string, 0, MaxReplayHeaderNames)
	for index := range MaxReplayHeaderNames {
		name := "X" + strings.Repeat("a", index)
		exactNames[name] = nil
		names = append(names, name)
	}
	if !replayHeadersFit(exactNames, names) {
		t.Fatal("replayHeadersFit() rejected exact header-name count")
	}
	names = append(names, "Overflow")
	if replayHeadersFit(exactNames, names) {
		t.Fatal("replayHeadersFit() accepted excessive header-name count")
	}

	exactName := strings.Repeat("A", MaxReplayHeaderNameBytes)
	if !replayHeadersFit(http.Header{exactName: {""}}, []string{exactName}) {
		t.Fatal("replayHeadersFit() rejected exact header-name length")
	}
	if replayHeadersFit(http.Header{"": {""}}, []string{""}) {
		t.Fatal("replayHeadersFit() accepted empty header name")
	}
	longName := exactName + "A"
	if replayHeadersFit(http.Header{longName: {""}}, []string{longName}) {
		t.Fatal("replayHeadersFit() accepted excessive header-name length")
	}

	exactValues := make([]string, MaxReplayHeaderValues)
	if !replayHeadersFit(http.Header{"A": nil, "B": exactValues}, []string{"A", "B"}) {
		t.Fatal("replayHeadersFit() rejected exact header-value count")
	}
	tooManyValues := make([]string, MaxReplayHeaderValues+1)
	if replayHeadersFit(http.Header{"A": tooManyValues}, []string{"A"}) {
		t.Fatal("replayHeadersFit() accepted excessive header-value count")
	}
	if !replayHeadersFit(http.Header{"A": {""}, "B": make([]string, MaxReplayHeaderValues-1)}, []string{"A", "B"}) {
		t.Fatal("replayHeadersFit() rejected exact split header-value count")
	}
	if replayHeadersFit(http.Header{"A": {""}, "B": make([]string, MaxReplayHeaderValues)}, []string{"A", "B"}) {
		t.Fatal("replayHeadersFit() accepted excessive split header-value count")
	}

	exactValue := strings.Repeat("x", MaxReplayHeaderValueBytes)
	if !replayHeadersFit(http.Header{"A": {exactValue}}, []string{"A"}) {
		t.Fatal("replayHeadersFit() rejected exact header-value length")
	}
	if replayHeadersFit(http.Header{"A": {exactValue + "x"}}, []string{"A"}) {
		t.Fatal("replayHeadersFit() accepted excessive header-value length")
	}

	values := make([]string, 8)
	remaining := MaxReplayHeaderBytes - len("A") - len("B")
	for index := range values {
		size := min(remaining, MaxReplayHeaderValueBytes)
		values[index] = strings.Repeat("x", size)
		remaining -= size
	}
	if !replayHeadersFit(http.Header{"A": values, "B": {""}}, []string{"A", "B"}) {
		t.Fatal("replayHeadersFit() rejected exact aggregate byte limit")
	}
	values[len(values)-1] += "x"
	if replayHeadersFit(http.Header{"A": values, "B": {""}}, []string{"A", "B"}) {
		t.Fatal("replayHeadersFit() accepted excessive aggregate bytes")
	}
}

func TestResponseSnapshotAndHeaderNamesHonorExactLimits(t *testing.T) {
	if !(responseSnapshot{Schema: replaySchema, Status: http.StatusOK, Body: make([]byte, MaxReplayResponseBytes)}).valid() {
		t.Fatal("responseSnapshot.valid() rejected exact body limit")
	}

	header := make(http.Header, 2*MaxReplayHeaderNames)
	for index := range 2 * MaxReplayHeaderNames {
		header["X"+strings.Repeat("a", index)] = nil
	}
	if names := headerNames(header); len(names) != MaxReplayHeaderNames+1 {
		t.Fatalf("headerNames() length = %d", len(names))
	}
}
