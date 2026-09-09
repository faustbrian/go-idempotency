// Package idempotencyrpc is the legacy durable JSON-RPC adapter.
//
// Deprecated: use github.com/faustbrian/go-idempotency/adapters/jsonrpc. This
// package remains supported for the longer of 180 days after successor
// availability and two subsequently published stable root-module minor
// releases.
package idempotencyrpc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/faustbrian/go-idempotency"
	canonical "github.com/faustbrian/go-idempotency/adapters/jsonrpc"
)

const (
	// MaxResponseBytes is the largest persisted JSON-RPC response envelope.
	MaxResponseBytes = canonical.MaxResponseBytes
	// MinResponseBytes leaves room for a durable internal-error response.
	MinResponseBytes = canonical.MinResponseBytes
)

// Request contains the business fields used for method-aware idempotency.
type Request struct {
	Method string
	Params json.RawMessage
}

// Error is a JSON-RPC protocol error that can be durably replayed.
type Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Response contains exactly one JSON result or protocol error.
type Response struct {
	Result json.RawMessage `json:"result,omitempty"`
	Error  *Error          `json:"error,omitempty"`
}

// Handler executes one elected JSON-RPC request owner.
type Handler func(context.Context, Request) Response

// KeyFunc constructs a caller and method-scoped key.
type KeyFunc func(context.Context, Request) (idempotency.Key, error)

// FingerprintFunc computes canonical business-request identity.
type FingerprintFunc func(Request) (idempotency.Fingerprint, error)

// Options configures method-aware durable JSON-RPC invocation.
type Options struct {
	Service           *idempotency.Service
	Lease             time.Duration
	MaxResponseBytes  int
	TransitionTimeout time.Duration
	Key               KeyFunc
	Fingerprint       FingerprintFunc
}

// CallResult reports the semantic outcome and any handler response.
type CallResult struct {
	Outcome  idempotency.Outcome
	Response Response
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
