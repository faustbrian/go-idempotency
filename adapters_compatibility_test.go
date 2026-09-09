package idempotency_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/faustbrian/go-idempotency"
	command "github.com/faustbrian/go-idempotency/adapters/command"
	httpadapter "github.com/faustbrian/go-idempotency/adapters/http"
	jsonrpc "github.com/faustbrian/go-idempotency/adapters/jsonrpc"
	otel "github.com/faustbrian/go-idempotency/adapters/otel"
	outbox "github.com/faustbrian/go-idempotency/adapters/outbox"
	queue "github.com/faustbrian/go-idempotency/adapters/queue"
	slogadapter "github.com/faustbrian/go-idempotency/adapters/slog"
	webhook "github.com/faustbrian/go-idempotency/adapters/webhook"
	legacycommand "github.com/faustbrian/go-idempotency/idempotencycommand"     //nolint:staticcheck // Legacy facade coverage.
	legacyhttp "github.com/faustbrian/go-idempotency/idempotencyhttp"           //nolint:staticcheck // Legacy facade coverage.
	legacylog "github.com/faustbrian/go-idempotency/idempotencylog"             //nolint:staticcheck // Legacy facade coverage.
	legacyoutbox "github.com/faustbrian/go-idempotency/idempotencyoutbox"       //nolint:staticcheck // Legacy facade coverage.
	legacyqueue "github.com/faustbrian/go-idempotency/idempotencyqueue"         //nolint:staticcheck // Legacy facade coverage.
	legacyrpc "github.com/faustbrian/go-idempotency/idempotencyrpc"             //nolint:staticcheck // Legacy facade coverage.
	legacytelemetry "github.com/faustbrian/go-idempotency/idempotencytelemetry" //nolint:staticcheck // Legacy facade coverage.
	legacywebhook "github.com/faustbrian/go-idempotency/idempotencywebhook"     //nolint:staticcheck // Legacy facade coverage.
	"github.com/faustbrian/go-idempotency/memory"
)

type adapterMessage struct{ payload []byte }

func (message adapterMessage) Payload() []byte { return message.payload }

type adapterClock struct{ now time.Time }

func (clock adapterClock) Now() time.Time { return clock.now }

var (
	_ func(command.Options) (*command.Runner, error)             = command.New
	_ func(httpadapter.Options) (*httpadapter.Middleware, error) = httpadapter.New
	_ func(jsonrpc.Options) (*jsonrpc.Middleware, error)         = jsonrpc.New
	_ func(queue.Options) (*queue.Middleware, error)             = queue.New
	_ func(webhook.Options) (*webhook.Processor, error)          = webhook.New

	_ func(legacycommand.Options) (*legacycommand.Runner, error)    = legacycommand.New
	_ func(legacyhttp.Options) (*legacyhttp.Middleware, error)      = legacyhttp.New
	_ func(legacyrpc.Options) (*legacyrpc.Middleware, error)        = legacyrpc.New
	_ func(legacyqueue.Options) (*legacyqueue.Middleware, error)    = legacyqueue.New
	_ func(legacywebhook.Options) (*legacywebhook.Processor, error) = legacywebhook.New

	_ outbox.Writer[adapterMessage]       = (legacyoutbox.Writer[adapterMessage])(nil)
	_ legacyoutbox.Writer[adapterMessage] = (outbox.Writer[adapterMessage])(nil)

	_ = otel.New
	_ = slogadapter.New
	_ = legacylog.New
	_ = legacytelemetry.New

	_ func(*queue.Middleware, func(context.Context, adapterMessage) error) func(context.Context, adapterMessage) error        = queue.Wrap[adapterMessage]
	_ func(*legacyqueue.Middleware, func(context.Context, adapterMessage) error) func(context.Context, adapterMessage) error  = legacyqueue.Wrap[adapterMessage]
	_ func(*webhook.Processor, func(context.Context, adapterMessage) error) func(context.Context, adapterMessage) error       = webhook.Wrap[adapterMessage]
	_ func(*legacywebhook.Processor, func(context.Context, adapterMessage) error) func(context.Context, adapterMessage) error = legacywebhook.Wrap[adapterMessage]
)

func TestCanonicalAdaptersAndLegacyFacadesRetainDistinctTypeIdentities(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		canonical     reflect.Type
		canonicalPath string
		legacy        reflect.Type
		legacyPath    string
	}{
		{"command", reflect.TypeOf((*command.Runner)(nil)).Elem(), "github.com/faustbrian/go-idempotency/adapters/command", reflect.TypeOf((*legacycommand.Runner)(nil)).Elem(), "github.com/faustbrian/go-idempotency/idempotencycommand"},
		{"http", reflect.TypeOf((*httpadapter.Middleware)(nil)).Elem(), "github.com/faustbrian/go-idempotency/adapters/http", reflect.TypeOf((*legacyhttp.Middleware)(nil)).Elem(), "github.com/faustbrian/go-idempotency/idempotencyhttp"},
		{"jsonrpc", reflect.TypeOf((*jsonrpc.Middleware)(nil)).Elem(), "github.com/faustbrian/go-idempotency/adapters/jsonrpc", reflect.TypeOf((*legacyrpc.Middleware)(nil)).Elem(), "github.com/faustbrian/go-idempotency/idempotencyrpc"},
		{"otel", reflect.TypeOf((*otel.Observer)(nil)).Elem(), "github.com/faustbrian/go-idempotency/adapters/otel", reflect.TypeOf((*legacytelemetry.Observer)(nil)).Elem(), "github.com/faustbrian/go-idempotency/idempotencytelemetry"},
		{"queue", reflect.TypeOf((*queue.Middleware)(nil)).Elem(), "github.com/faustbrian/go-idempotency/adapters/queue", reflect.TypeOf((*legacyqueue.Middleware)(nil)).Elem(), "github.com/faustbrian/go-idempotency/idempotencyqueue"},
		{"slog", reflect.TypeOf((*slogadapter.Observer)(nil)).Elem(), "github.com/faustbrian/go-idempotency/adapters/slog", reflect.TypeOf((*legacylog.Observer)(nil)).Elem(), "github.com/faustbrian/go-idempotency/idempotencylog"},
		{"webhook", reflect.TypeOf((*webhook.Processor)(nil)).Elem(), "github.com/faustbrian/go-idempotency/adapters/webhook", reflect.TypeOf((*legacywebhook.Processor)(nil)).Elem(), "github.com/faustbrian/go-idempotency/idempotencywebhook"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := test.canonical.PkgPath(); got != test.canonicalPath {
				t.Fatalf("canonical PkgPath() = %q, want %q", got, test.canonicalPath)
			}
			if got := test.legacy.PkgPath(); got != test.legacyPath {
				t.Fatalf("legacy PkgPath() = %q, want %q", got, test.legacyPath)
			}
		})
	}
}

func TestLegacyAdapterErrorsRemainCanonicalSentinels(t *testing.T) {
	t.Parallel()

	if legacycommand.ErrInProgress != command.ErrInProgress ||
		legacycommand.ErrConflict != command.ErrConflict ||
		legacycommand.ErrTerminalFailure != command.ErrTerminalFailure {
		t.Fatal("command facade does not expose canonical error sentinels")
	}
	if legacyhttp.ErrResponseTooLarge != httpadapter.ErrResponseTooLarge {
		t.Fatal("HTTP facade does not expose canonical error sentinel")
	}
	if legacylog.ErrNilLogger != slogadapter.ErrNilLogger {
		t.Fatal("slog facade does not expose canonical error sentinel")
	}
	if legacyqueue.ErrInProgress != queue.ErrInProgress ||
		legacyqueue.ErrConflict != queue.ErrConflict ||
		legacyqueue.ErrTerminalFailure != queue.ErrTerminalFailure {
		t.Fatal("queue facade does not expose canonical error sentinels")
	}
	if legacytelemetry.ErrNilMeterProvider != otel.ErrNilMeterProvider {
		t.Fatal("OpenTelemetry facade does not expose canonical error sentinel")
	}
	if legacywebhook.ErrInProgress != webhook.ErrInProgress ||
		legacywebhook.ErrConflict != webhook.ErrConflict ||
		legacywebhook.ErrTerminalFailure != webhook.ErrTerminalFailure {
		t.Fatal("webhook facade does not expose canonical error sentinels")
	}
}

func TestLegacyJSONRPCFacadePreservesFreshResponseIdentity(t *testing.T) {
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
	middleware, err := legacyrpc.New(legacyrpc.Options{
		Service: service,
		Lease:   time.Minute,
		Key: func(context.Context, legacyrpc.Request) (idempotency.Key, error) {
			return idempotency.NewKey("rpc", "tenant", "widgets.get", "caller", "request-1")
		},
		Fingerprint: func(legacyrpc.Request) (idempotency.Fingerprint, error) {
			return fingerprint, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	protocolError := &legacyrpc.Error{Code: -32001, Message: "rejected"}
	result, err := middleware.Call(
		context.Background(),
		legacyrpc.Request{Method: "widgets.get", Params: []byte(`{"id":1}`)},
		func(context.Context, legacyrpc.Request) legacyrpc.Response {
			return legacyrpc.Response{Error: protocolError}
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Response.Error != protocolError {
		t.Fatal("fresh facade response replaced the handler's protocol error identity")
	}
}

func TestLegacyWebhookFacadeValidatesNilHandlerWithoutProcessor(t *testing.T) {
	t.Parallel()

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("Wrap(nil, nil) panicked: %v", recovered)
		}
	}()

	wrapped := legacywebhook.Wrap[adapterMessage](nil, nil)
	err := wrapped(context.Background(), adapterMessage{})
	var idempotencyError *idempotency.Error
	if !errors.As(err, &idempotencyError) ||
		idempotencyError.Reason != idempotency.ReasonInvalidConfiguration ||
		idempotencyError.Field != "handler" {
		t.Fatalf("Wrap(nil, nil) error = %v, want invalid handler configuration", err)
	}
}
