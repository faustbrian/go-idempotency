# Webhook delivery deduplication

`idempotencywebhook` executes a verified provider delivery once and treats a
completed redelivery as successfully handled.

Verify the provider signature and bound the request body before entering the
processor. Never persist or act on an unauthenticated delivery merely to dedupe
it.

```go
processor, err := idempotencywebhook.New(idempotencywebhook.Options{
	Service: service,
	Lease: 30 * time.Second,
	TransitionTimeout: 5 * time.Second,
	Key: func(ctx context.Context, delivery idempotencywebhook.Delivery) (idempotency.Key, error) {
		provider, ok := delivery.(ProviderDelivery)
		if !ok {
			return idempotency.Key{}, &idempotency.Error{
				Reason: idempotency.ReasonInvalidPayload,
				Field: "provider_delivery",
			}
		}
		return idempotency.NewKey(
			"webhooks",
			provider.AccountID(),
			provider.ProviderName(),
			"orders-endpoint",
			provider.DeliveryID(),
		)
	},
	Fingerprint: func(delivery idempotencywebhook.Delivery) (idempotency.Fingerprint, error) {
		return canonical.BytesFingerprint(
			"provider-wire-payload-v1",
			delivery.Payload(),
			256*1024,
		)
	},
})
if err != nil {
	return err
}

handle := idempotencywebhook.Wrap(
	processor,
	func(ctx context.Context, delivery ProviderDelivery) error {
		ownership, _ := idempotency.OwnershipFromContext(ctx)
		return applyEventWithFence(ctx, delivery, ownership.FencingToken)
	},
)
```

The provider name belongs in the operation and the provider's immutable
delivery or event ID belongs in the key value. Scope by destination account and
endpoint or consumer identity. Do not use request arrival time, signature,
retry count, or a load balancer request ID.

Fingerprint the exact verified bytes when the provider signs and retries those
bytes. If semantically equivalent encodings are expected, use a documented
provider-specific canonicalization policy instead. A changed payload under the
same provider delivery ID returns `ErrConflict` for investigation.

## Provider response mapping

- completed replay returns `nil`; acknowledge with the provider's success code;
- handler failure releases ownership and should return the provider's retryable
  response;
- `ErrInProgress` should return a retryable response or short bounded backoff;
- `ErrConflict` should alert and avoid blindly applying either payload;
- `ErrTerminalFailure` follows the application's permanent rejection policy;
- storage failure fails closed and must not be acknowledged as processed.

The handler context carries fencing ownership. A provider delivery ID prevents
ordinary redelivery duplicates, but it does not stop an expired old handler
from committing late.

The `go-webhook` repository currently has no module or public delivery API.
`Delivery` therefore uses the minimal structural `Payload() []byte` contract and
`Wrap` preserves an application's concrete delivery type. Add the direct typed
binding when that repository publishes stable provider interfaces.
