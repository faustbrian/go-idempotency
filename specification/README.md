# Specification conformance matrix

The [specification decision register](../docs/specification-decisions.md) owns
interpretation and policy. This matrix binds every decision once to its primary
observable evidence; the machine-complete bindings are in
[conformance.json](conformance.json).

| Decision | Public behavior | Evidence boundary |
| --- | --- | --- |
| IDEMPOTENCY-DEC-001 | Hostile or ambiguous JSON fails before canonical identity is returned. | `TestJSONRejectsHostileOrAmbiguousInput`, `TestJSONAcceptsEscapesAndSurrogateRangeBoundaries`, `canonical/testdata/rfc8785.json`, `FuzzJSONIsIdempotent`, `scripts/check-jcs-differential.mjs` |
| IDEMPOTENCY-DEC-002 | Accepted binary64 values use JCS spelling while negative zero is rejected. | `TestJSONUsesRFC8785CanonicalForm`, `TestJSONRejectsHostileOrAmbiguousInput`, `TestJSONAcceptsPositiveZeroAndNegativeNonzero`, `canonical/testdata/rfc8785.json`, `FuzzJSONIsIdempotent`, `scripts/check-jcs-differential.mjs` |
| IDEMPOTENCY-DEC-003 | Properties use recursive UTF-16 ordering without Unicode normalization. | `TestJSONMatchesPinnedRFC8785Fixtures`, `TestJSONAcceptsValidEscapesAndSurrogatePairs`, `canonical/testdata/rfc8785.json`, `FuzzJSONIsIdempotent`, `scripts/check-jcs-differential.mjs` |
| IDEMPOTENCY-DEC-004 | Canonical and raw-byte fingerprint work is explicitly bounded. | `TestJSONEnforcesAllResourceLimits`, `TestJSONAcceptsExactResourceLimits`, `TestBytesFingerprintAcceptsExactBound`, `FuzzBytesFingerprintPreservesEncoding` |
| IDEMPOTENCY-DEC-005 | Fingerprints use versioned SHA-256 and correlations use framed HMAC-SHA-256. | `TestFingerprintUsesAnExplicitVersion`, `TestFingerprintCanBeReconstructedFromPersistedDigest`, `TestNewHMACKeyHasherProtectsLogicalIdentity`, `FuzzFingerprintPolicyVersionsRemainDistinct` |
| IDEMPOTENCY-DEC-006 | JSON-RPC durable identity projects method and params, not transport id. | `TestMiddlewareValidatesInvocationIdentity`, `TestMiddlewareExecutesOnceAndReplaysResult` |
| IDEMPOTENCY-DEC-007 | Exactly one valid result or error projection is durably replayed. | `TestMiddlewareExecutesOnceAndReplaysResult`, `TestMiddlewareReplaysJSONRPCError`, `TestMiddlewareFailsClosedAtStorageAndPersistedReplayBoundaries`, `FuzzMalformedReplayFailsClosed` |
| IDEMPOTENCY-DEC-008 | Unreplayable handler output becomes terminal JSON-RPC -32603. | `TestMiddlewareRecordsInvalidOrOversizedHandlerResponseAsTerminal`, `TestMiddlewareReplaysJSONRPCError` |
| IDEMPOTENCY-DEC-009 | Notification and batch transport semantics remain outside middleware. | `TestMiddlewareValidatesInvocationIdentity`, `TestMiddlewareReturnsConflictAndInProgressWithoutExecuting` |
| IDEMPOTENCY-DEC-010 | The application validates and scopes one nonempty trimmed key value. | `TestMiddlewareRequiresIdempotencyKey`, `TestMiddlewareRejectsKeyAndFingerprintFailures`, `TestNewKeyRequiresEveryIdentityPart` |
| IDEMPOTENCY-DEC-011 | Distinct conflict outcomes share stable HTTP 409 in the v1 profile. | `TestMiddlewareReturnsExplicitProtocolOutcomes`, `TestMiddlewareRejectsKeyAndFingerprintFailures`, `TestBeginFailsClosedWhenStorageIsUnavailable` |
| IDEMPOTENCY-DEC-012 | Bounded response projections replay and unreplayable output becomes terminal. | `TestMiddlewareExecutesOnceAndReplaysResponse`, `TestMiddlewareBoundsHandlerResponseAndRecordsTerminalFailure`, `TestMiddlewareDeduplicatesReplayHeaders`, `FuzzMalformedReplayFailsClosed` |
| IDEMPOTENCY-DEC-013 | Durable ownership and fencing prevent leases from implying exactly-once execution. | `TestBeginMapsDurableOutcomesToExecutionDecisions`, `TestStoreConformance`, `TestBeginFailsClosedWhenStorageIsUnavailable` |

## Claim boundaries

- RFC 8785 accepted cases are checked against pinned official fixtures and the
  maintained `canonicalize` 4.0.0 peer. Negative zero is an explicit deliberate
  policy difference, not a conformance claim.
- JSON-RPC evidence covers the package's per-invocation durable projection. It
  does not claim full request parsing, notification suppression, batch assembly,
  transport IDs, or server interoperability.
- The HTTP adapter is compared with draft-07 but does not claim full draft
  conformance: structured-field parsing and the recommended 422 mapping are not
  implemented by the stable v1 profile.
- Durable ownership, expiry, crash recovery, and fencing are package policy.
  They do not establish exactly-once execution.

Run `make conformance` for focused Go evidence and maintained-peer JCS
comparison. CI additionally runs the canonical online source and errata check.
