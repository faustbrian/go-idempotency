// Package idempotencyrpc is the legacy durable JSON-RPC adapter.
// Dependencies are borrowed; callbacks are synchronous trusted application code.
// No automatic lease renewal, worker/client closure or exactly-once effect is owned.
// Callers enforce authentication, fencing, request admission and IO deadlines.
//
// Deprecated: use github.com/faustbrian/go-idempotency/v2/adapters/jsonrpc. This
// package remains supported for the longer of 180 days after successor
// availability and two subsequently published stable root-module minor
// releases.
package idempotencyrpc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/faustbrian/go-idempotency/v2"
	canonical "github.com/faustbrian/go-idempotency/v2/adapters/jsonrpc"
)

const (
	// MaxResponseBytes is the largest persisted JSON-RPC response envelope.
	MaxResponseBytes = canonical.MaxResponseBytes
	// MinResponseBytes leaves room for a durable internal-error response.
	MinResponseBytes = canonical.MinResponseBytes
)

// Request contains the business fields used for method-aware idempotency.
type Request struct {
	// Method is nonempty and must exactly match the constructed key operation.
	Method string
	// Params is borrowed business input; transport ids/batches/notifications are not
	// modeled.
	Params json.RawMessage
}

// Error is a JSON-RPC protocol error that can be durably replayed.
type Error struct {
	// Code is the application-selected integer protocol error code.
	Code int `json:"code"`
	// Message must be nonempty for admitted replay; it is not secret-filtered.
	Message string `json:"message"`
	// Data is optional valid JSON, retained without business-schema validation.
	Data json.RawMessage `json:"data,omitempty"`
}

// Response contains exactly one JSON result or protocol error.
type Response struct {
	// Result is valid JSON when Error is nil; empty is not a valid result.
	Result json.RawMessage `json:"result,omitempty"`
	// Error is the alternative protocol response; both/neither arms are invalid.
	Error *Error `json:"error,omitempty"`
}

// Handler executes one elected JSON-RPC request owner.
type Handler func(context.Context, Request) Response

// KeyFunc constructs a caller and method-scoped key.
type KeyFunc func(context.Context, Request) (idempotency.Key, error)

// FingerprintFunc computes canonical business-request identity.
type FingerprintFunc func(Request) (idempotency.Fingerprint, error)

// Options configures method-aware durable JSON-RPC invocation.
type Options struct {
	// Service is borrowed; its store determines actual durability and concurrency.
	Service *idempotency.Service
	// Lease is positive and at most MaxLease; handlers do not receive automatic renewal.
	Lease time.Duration
	// MaxResponseBytes caps encoded replay after encoding; zero selects 64 KiB, range is
	// 256 to 900 KiB.
	MaxResponseBytes int
	// TransitionTimeout supplies a detached cleanup context deadline; zero selects five
	// seconds.
	TransitionTimeout time.Duration
	// Key is trusted application scoping/authentication policy, called synchronously.
	Key KeyFunc
	// Fingerprint is trusted canonical business identity policy, called synchronously.
	Fingerprint FingerprintFunc
}

// CallResult reports the semantic outcome and any handler response.
type CallResult struct {
	// Outcome classifies acquisition or terminal response encoding failure.
	Outcome idempotency.Outcome
	// Response is application response data; fresh raw bytes/pointers can still alias
	// handler output.
	Response Response
	// Replayed reports decoded retained data rather than a fresh handler call.
	Replayed bool
}

// Middleware preserves the legacy JSON-RPC adapter type identity.
type Middleware struct{ inner *canonical.Middleware }

// New validates options and constructs JSON-RPC middleware.
func New(options Options) (*Middleware, error) {
	var key canonical.KeyFunc
	if options.Key != nil {
		key = func(ctx context.Context, request canonical.Request) (idempotency.Key, error) {
			return options.Key(ctx, Request(request))
		}
	}
	var fingerprint canonical.FingerprintFunc
	if options.Fingerprint != nil {
		fingerprint = func(request canonical.Request) (idempotency.Fingerprint, error) {
			return options.Fingerprint(Request(request))
		}
	}
	inner, err := canonical.New(canonical.Options{
		Service: options.Service, Lease: options.Lease,
		MaxResponseBytes:  options.MaxResponseBytes,
		TransitionTimeout: options.TransitionTimeout,
		Key:               key, Fingerprint: fingerprint,
	})
	if err != nil {
		return nil, err
	}

	return &Middleware{inner: inner}, nil
}

// Call invokes handler only for a newly acquired or taken-over request.
func (middleware *Middleware) Call(
	ctx context.Context,
	request Request,
	handler Handler,
) (CallResult, error) {
	var next canonical.Handler
	var fresh Response
	handlerCalled := false
	if handler != nil {
		next = func(ctx context.Context, request canonical.Request) canonical.Response {
			handlerCalled = true
			fresh = handler(ctx, Request(request))

			return canonicalResponse(fresh)
		}
	}
	result, err := middleware.inner.Call(ctx, canonical.Request(request), next)
	response := legacyResponse(result.Response)
	if err == nil && handlerCalled && result.Outcome != idempotency.OutcomeTerminalFailure {
		response = fresh
	}

	return CallResult{
		Outcome: result.Outcome, Response: response, Replayed: result.Replayed,
	}, err
}

func canonicalResponse(response Response) canonical.Response {
	var protocolError *canonical.Error
	if response.Error != nil {
		protocolError = &canonical.Error{
			Code: response.Error.Code, Message: response.Error.Message, Data: response.Error.Data,
		}
	}

	return canonical.Response{Result: response.Result, Error: protocolError}
}

func legacyResponse(response canonical.Response) Response {
	var protocolError *Error
	if response.Error != nil {
		protocolError = &Error{
			Code: response.Error.Code, Message: response.Error.Message, Data: response.Error.Data,
		}
	}

	return Response{Result: response.Result, Error: protocolError}
}
