package idempotency_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/faustbrian/go-idempotency/v3"
	jsonrpc "github.com/faustbrian/go-idempotency/v3/adapters/jsonrpc"
	webhook "github.com/faustbrian/go-idempotency/v3/adapters/webhook"
	"github.com/faustbrian/go-idempotency/v3/memory"
)

type adapterMessage struct{ payload []byte }

func (message adapterMessage) Payload() []byte { return message.payload }

type adapterClock struct{ now time.Time }

func (clock adapterClock) Now() time.Time { return clock.now }

func TestJSONRPCAdapterPreservesFreshResponseIdentity(t *testing.T) {
	t.Parallel()

	store, err := memory.New(memory.Options{
		Clock: adapterClock{now: time.Unix(1_700_000_000, 0)},
		OwnerTokens: func() (string, error) {
			return "owner", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	service, err := idempotency.NewService(store)
	if err != nil {
		t.Fatal(err)
	}
	fingerprint, err := idempotency.NewFingerprint("rpc-v1", []byte(`{"id":1}`))
	if err != nil {
		t.Fatal(err)
	}
	middleware, err := jsonrpc.New(jsonrpc.Options{
		Service: service,
		Lease:   time.Minute,
		Key: func(context.Context, jsonrpc.Request) (idempotency.Key, error) {
			return idempotency.NewKey("rpc", "tenant", "widgets.get", "caller", "request-1")
		},
		Fingerprint: func(jsonrpc.Request) (idempotency.Fingerprint, error) {
			return fingerprint, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	protocolError := &jsonrpc.Error{Code: -32001, Message: "rejected"}
	result, err := middleware.Call(
		context.Background(),
		jsonrpc.Request{Method: "widgets.get", Params: []byte(`{"id":1}`)},
		func(context.Context, jsonrpc.Request) jsonrpc.Response {
			return jsonrpc.Response{Error: protocolError}
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Response.Error != protocolError {
		t.Fatal("fresh adapter response replaced the handler's protocol error identity")
	}
}

func TestWebhookAdapterValidatesNilHandlerWithoutProcessor(t *testing.T) {
	t.Parallel()

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("Wrap(nil, nil) panicked: %v", recovered)
		}
	}()

	wrapped := webhook.Wrap[adapterMessage](nil, nil)
	err := wrapped(context.Background(), adapterMessage{})
	var idempotencyError *idempotency.Error
	if !errors.As(err, &idempotencyError) ||
		idempotencyError.Reason != idempotency.ReasonInvalidConfiguration ||
		idempotencyError.Field != "handler" {
		t.Fatalf("Wrap(nil, nil) error = %v, want invalid handler configuration", err)
	}
}
