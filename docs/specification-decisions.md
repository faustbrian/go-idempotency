# Idempotency specification decisions

This register records every material standards interpretation and package-owned
policy for canonical fingerprints, JSON-RPC projection, HTTP idempotency, and
durable enforcement. Exact source bytes and change authorities are pinned in
the [source manifest](../specification/manifest.tsv) and
[monitoring policy](../specification/monitoring.json). Passing tests or peer
agreement does not replace the selected behavior recorded here.

Statuses are `resolved`, `unresolved`, or `superseded`. Observable decision
changes require compatibility, wire, security, resource, changelog, executable
evidence, and history review.

## IDEMPOTENCY-DEC-001: JCS input domain and hostile JSON rejection

- **Status:** resolved
- **Owner:** idempotency maintainers
- **Classification:** interoperability policy
- **Decision scope:** normative
- **Specification:** RFC 8785 JSON Canonicalization Scheme
- **Version:** RFC 8785
- **Source authority:** rfc8785-source
- **Section:** 3.1
- **Requirement strength:** MUST
- **Issue:** JCS requires I-JSON input but generic Go JSON decoding can accept or collapse representations that are unsafe for request identity.
- **Peer behavior:** The maintained canonicalize 4.0.0 peer and the production JCS implementation agree on accepted pinned RFC fixtures.
- **Selected behavior:** JSON rejects malformed UTF-8, lone or malformed surrogates, duplicate names, trailing values, and numbers outside binary64 before returning canonical bytes.
- **Rationale:** Ambiguous input cannot safely identify a retried business request.
- **Security consequences:** Duplicate-name and Unicode parser differentials fail closed.
- **Resource consequences:** Validation occurs under explicit input and depth limits.
- **Compatibility consequences:** Inputs accepted by permissive JSON parsers can be rejected.
- **Wire consequences:** Only one complete UTF-8 JSON value enters canonicalization.
- **Upstream status:** RFC 8785 errata are monitored separately.
- **Reconsider when:** RFC 8785, its errata, or the accepted input policy changes.
- **Authoritative URL:** https://www.rfc-editor.org/rfc/rfc8785.txt
- **Credible interpretation:** Accept whatever encoding/json decodes
- **Credible interpretation:** Let duplicate object names select a value
- **Credible interpretation:** Require valid UTF-8, paired surrogates, unique names, and one complete JSON value
- **Executable evidence:** TestJSONRejectsHostileOrAmbiguousInput
- **Executable evidence:** TestJSONAcceptsEscapesAndSurrogateRangeBoundaries
- **Fixture evidence:** canonical/testdata/rfc8785.json
- **Fuzz evidence:** FuzzJSONIsIdempotent
- **Interoperability evidence:** canonical/testdata/rfc8785.json
- **Differential evidence:** scripts/check-jcs-differential.mjs
- **Public API:** canonical.JSON
- **Public API:** canonical.JSONFingerprint
- **Documentation:** docs/specification-decisions.md
- **Documentation:** docs/fingerprints.md
- **Additional authoritative source:** `{"id":"rfc8259-source","version":"RFC 8259","url":"https://www.rfc-editor.org/rfc/rfc8259.txt","specifications":["RFC 8259 JSON"]}`

## IDEMPOTENCY-DEC-002: JCS number serialization and negative zero policy

- **Status:** resolved
- **Owner:** idempotency maintainers
- **Classification:** interoperability policy
- **Decision scope:** defensive
- **Specification:** RFC 8785 JSON Canonicalization Scheme
- **Version:** RFC 8785
- **Source authority:** rfc8785-source
- **Section:** 3.2.2.3 and Appendix B
- **Requirement strength:** not specified
- **Issue:** RFC 8785 serializes binary64 values using ECMAScript rules and maps minus zero to zero, while request identity can treat a signed zero as suspicious input.
- **Peer behavior:** canonicalize 4.0.0 serializes negative zero as 0 as required by RFC 8785.
- **Selected behavior:** Finite binary64 numbers use RFC 8785 serialization, but negative zero is deliberately rejected as invalid payload.
- **Rationale:** Rejecting negative zero prevents a source-level distinction from silently collapsing in business identity.
- **Security consequences:** Signed-zero ambiguity fails closed.
- **Resource consequences:** Number conversion is bounded to binary64.
- **Compatibility consequences:** This defensive policy is stricter than RFC 8785 for negative zero.
- **Wire consequences:** Accepted numbers emit ECMAScript-compatible shortest JSON spellings.
- **Upstream status:** The negative-zero difference is deliberate package policy, not an RFC erratum.
- **Reconsider when:** A new fingerprint policy version explicitly adopts RFC 8785 negative-zero equivalence.
- **Authoritative URL:** https://www.rfc-editor.org/rfc/rfc8785.txt
- **Credible interpretation:** Serialize negative zero as 0
- **Credible interpretation:** Preserve the source spelling
- **Credible interpretation:** Reject negative zero before JCS serialization
- **Executable evidence:** TestJSONUsesRFC8785CanonicalForm
- **Executable evidence:** TestJSONRejectsHostileOrAmbiguousInput
- **Executable evidence:** TestJSONAcceptsPositiveZeroAndNegativeNonzero
- **Fixture evidence:** canonical/testdata/rfc8785.json
- **Fuzz evidence:** FuzzJSONIsIdempotent
- **Interoperability evidence:** canonical/testdata/rfc8785.json
- **Differential evidence:** scripts/check-jcs-differential.mjs
- **Public API:** canonical.JSON
- **Public API:** canonical.JSONFingerprint
- **Documentation:** docs/specification-decisions.md
- **Documentation:** docs/fingerprints.md

## IDEMPOTENCY-DEC-003: JCS property ordering and Unicode preservation

- **Status:** resolved
- **Owner:** idempotency maintainers
- **Classification:** implementation-defined behavior
- **Decision scope:** normative
- **Specification:** RFC 8785 JSON Canonicalization Scheme
- **Version:** RFC 8785
- **Source authority:** rfc8785-source
- **Section:** 3.2.3 and 3.2.4
- **Requirement strength:** MUST
- **Issue:** UTF-8 byte order, Unicode scalar order, locale order, and RFC-required UTF-16 code-unit order can differ for object property names.
- **Peer behavior:** The maintained canonicalize 4.0.0 peer agrees on the pinned RFC UTF-16 property-order fixture.
- **Selected behavior:** Objects are recursively sorted by raw UTF-16 property-name code units, arrays retain element order, and output is UTF-8.
- **Rationale:** The RFC order is required for cross-runtime fingerprint agreement.
- **Security consequences:** Locale and encoding-dependent property ordering cannot alter fingerprints.
- **Resource consequences:** Sorting remains bounded by caller input, output, and depth limits.
- **Compatibility consequences:** UTF-8-sorted encoders can produce incompatible fingerprints.
- **Wire consequences:** Unicode strings are preserved without normalization while property order is deterministic.
- **Upstream status:** No package-specific exception to RFC ordering is selected.
- **Reconsider when:** A superseding canonical JSON profile changes ordering or Unicode treatment.
- **Authoritative URL:** https://www.rfc-editor.org/rfc/rfc8785.txt
- **Credible interpretation:** Sort UTF-8 bytes
- **Credible interpretation:** Sort Unicode scalar values
- **Credible interpretation:** Sort raw property names by UTF-16 code units recursively
- **Executable evidence:** TestJSONMatchesPinnedRFC8785Fixtures
- **Executable evidence:** TestJSONAcceptsValidEscapesAndSurrogatePairs
- **Fixture evidence:** canonical/testdata/rfc8785.json
- **Fuzz evidence:** FuzzJSONIsIdempotent
- **Interoperability evidence:** canonical/testdata/rfc8785.json
- **Differential evidence:** scripts/check-jcs-differential.mjs
- **Public API:** canonical.JSON
- **Public API:** canonical.JSONFingerprint
- **Documentation:** docs/specification-decisions.md
- **Documentation:** docs/fingerprints.md

## IDEMPOTENCY-DEC-004: Canonicalization bounds and raw-byte policy

- **Status:** resolved
- **Owner:** idempotency maintainers
- **Classification:** omission
- **Decision scope:** application-policy
- **Specification:** RFC 8785 JSON Canonicalization Scheme
- **Version:** RFC 8785
- **Source authority:** rfc8785-source
- **Section:** 3
- **Requirement strength:** not specified
- **Issue:** RFC 8785 defines output bytes but does not define application resource limits or when already-canonical non-JSON bytes may be fingerprinted.
- **Peer behavior:** General JCS peers do not own this package's hostile-input resource policy or non-JSON encoding contracts.
- **Selected behavior:** JSON requires positive input, output, and depth limits; BytesFingerprint preserves exact bytes and requires a positive maximum.
- **Rationale:** Callers must choose the resource and encoding policy that defines business identity.
- **Security consequences:** Untrusted canonicalization and hashing are bounded before authority is granted.
- **Resource consequences:** Exact configured boundaries are accepted and the next byte or level is rejected.
- **Compatibility consequences:** Changing limits or raw encoding requires application review.
- **Wire consequences:** BytesFingerprint performs no normalization or transcoding.
- **Upstream status:** Resource and raw-byte policy are package-owned.
- **Reconsider when:** A public profile defines different mandatory bounds or a non-JSON canonical encoding.
- **Authoritative URL:** https://www.rfc-editor.org/rfc/rfc8785.txt
- **Credible interpretation:** Canonicalize without limits
- **Credible interpretation:** Apply hidden defaults
- **Credible interpretation:** Require explicit input, output, depth, and raw-byte limits
- **Executable evidence:** TestJSONEnforcesAllResourceLimits
- **Executable evidence:** TestJSONAcceptsExactResourceLimits
- **Executable evidence:** TestBytesFingerprintAcceptsExactBound
- **Fuzz evidence:** FuzzBytesFingerprintPreservesEncoding
- **Public API:** canonical.Limits
- **Public API:** canonical.JSON
- **Public API:** canonical.JSONFingerprint
- **Public API:** canonical.BytesFingerprint
- **Documentation:** docs/specification-decisions.md
- **Documentation:** docs/fingerprints.md
- **Documentation:** docs/resource-budgets.md

## IDEMPOTENCY-DEC-005: Fingerprint and correlation digest framing

- **Status:** resolved
- **Owner:** idempotency maintainers
- **Classification:** omission
- **Decision scope:** defensive
- **Specification:** RFC 6234 SHA Algorithms
- **Version:** RFC 6234
- **Source authority:** rfc6234-source
- **Section:** 6 and 8
- **Requirement strength:** not specified
- **Issue:** SHA-256 and HMAC define primitives but not fingerprint policy versions or collision-safe framing of logical key components.
- **Peer behavior:** Standard-library SHA-256 and HMAC implement the pinned RFC primitives; no external peer selects package framing.
- **Selected behavior:** Fingerprints are a policy version plus SHA-256 of canonical bytes; diagnostic correlation is HMAC-SHA-256 over five length-prefixed key components.
- **Rationale:** Versions prevent silent canonicalization migration and length prefixes prevent concatenation ambiguity.
- **Security consequences:** Correlation values do not expose raw logical keys and secrets require at least 32 bytes.
- **Resource consequences:** Versions, key parts, and fixed-width digests are bounded.
- **Compatibility consequences:** An identical digest under a different policy version is a conflict.
- **Wire consequences:** Persisted fingerprint sums are exactly 32 bytes and correlation output is lowercase hexadecimal.
- **Upstream status:** RFC 2104 supplies the HMAC construction as an additional authority.
- **Reconsider when:** A versioned digest migration or correlation format is introduced.
- **Authoritative URL:** https://www.rfc-editor.org/rfc/rfc6234.txt
- **Credible interpretation:** Hash unframed concatenation
- **Credible interpretation:** Hash canonical bytes without a policy version
- **Credible interpretation:** Use SHA-256 fingerprints plus version comparison and length-framed HMAC-SHA-256 correlation
- **Executable evidence:** TestFingerprintUsesAnExplicitVersion
- **Executable evidence:** TestFingerprintCanBeReconstructedFromPersistedDigest
- **Executable evidence:** TestNewHMACKeyHasherProtectsLogicalIdentity
- **Fuzz evidence:** FuzzFingerprintPolicyVersionsRemainDistinct
- **Public API:** Fingerprint
- **Public API:** NewFingerprint
- **Public API:** NewFingerprintFromSum
- **Public API:** NewHMACKeyHasher
- **Documentation:** docs/specification-decisions.md
- **Documentation:** docs/fingerprints.md
- **Documentation:** docs/observability.md
- **Additional authoritative source:** `{"id":"rfc2104-source","version":"RFC 2104","url":"https://www.rfc-editor.org/rfc/rfc2104.txt","specifications":["RFC 2104 HMAC"]}`

## IDEMPOTENCY-DEC-006: JSON-RPC invocation projection and identity

- **Status:** resolved
- **Owner:** idempotency maintainers
- **Classification:** omission
- **Decision scope:** transport-specific
- **Specification:** JSON-RPC 2.0 Specification
- **Version:** 2.0 updated 2013-01-04
- **Source authority:** jsonrpc20-source
- **Section:** 4 and 4.2
- **Requirement strength:** not specified
- **Issue:** JSON-RPC defines full request objects but does not define which fields establish durable business idempotency identity.
- **Peer behavior:** Full JSON-RPC servers bind transport IDs to responses but do not provide this package's durable business-key projection.
- **Selected behavior:** Request projects only method and params; the application key operation must exactly equal method and the JSON-RPC id is not used as the business idempotency key.
- **Rationale:** Transport correlation IDs are not necessarily stable, unique business identities.
- **Security consequences:** Tenant, caller, namespace, operation, and delivery identity remain explicit application inputs.
- **Resource consequences:** Params fingerprinting is caller-owned and must be bounded.
- **Compatibility consequences:** Changing projected business fields requires a new fingerprint policy version.
- **Wire consequences:** jsonrpc and id remain owned by the surrounding transport adapter.
- **Upstream status:** JSON-RPC 2.0 has no separate errata feed; its publication index is monitored.
- **Reconsider when:** A transport adapter standardizes durable key projection.
- **Authoritative URL:** https://www.jsonrpc.org/specification
- **Credible interpretation:** Use the JSON-RPC id alone
- **Credible interpretation:** Hash the complete transport envelope
- **Credible interpretation:** Project method and params while requiring an application-scoped key
- **Executable evidence:** TestMiddlewareValidatesInvocationIdentity
- **Executable evidence:** TestMiddlewareExecutesOnceAndReplaysResult
- **Public API:** idempotencyrpc.Request
- **Public API:** idempotencyrpc.KeyFunc
- **Public API:** idempotencyrpc.FingerprintFunc
- **Public API:** idempotencyrpc.Middleware.Call
- **Documentation:** docs/specification-decisions.md
- **Documentation:** docs/json-rpc.md
- **Additional authoritative source:** `{"id":"rfc8259-source","version":"RFC 8259","url":"https://www.rfc-editor.org/rfc/rfc8259.txt","specifications":["RFC 8259 JSON"]}`

## IDEMPOTENCY-DEC-007: JSON-RPC result and error replay shape

- **Status:** resolved
- **Owner:** idempotency maintainers
- **Classification:** interoperability policy
- **Decision scope:** normative
- **Specification:** JSON-RPC 2.0 Specification
- **Version:** 2.0 updated 2013-01-04
- **Source authority:** jsonrpc20-source
- **Section:** 5 and 5.1
- **Requirement strength:** MUST
- **Issue:** Durable replay must preserve the JSON-RPC result-or-error invariant without pretending to own the full response envelope.
- **Peer behavior:** JSON-RPC 2.0 peers require result and error to be mutually exclusive.
- **Selected behavior:** Response contains exactly one valid JSON result or an error with an integer code, nonempty message, and optional valid JSON data.
- **Rationale:** Replay must not manufacture a protocol-invalid semantic response.
- **Security consequences:** Malformed persisted responses fail closed before handler execution.
- **Resource consequences:** The versioned persisted response envelope is size-bounded.
- **Compatibility consequences:** Full jsonrpc and id members remain transport-owned and are not persisted here.
- **Wire consequences:** Result and error JSON values are replayed through typed projection fields.
- **Upstream status:** The projection follows the current JSON-RPC 2.0 result and error invariant.
- **Reconsider when:** The adapter begins owning complete JSON-RPC response envelopes.
- **Authoritative URL:** https://www.jsonrpc.org/specification
- **Credible interpretation:** Store arbitrary response objects
- **Credible interpretation:** Allow both result and error
- **Credible interpretation:** Store exactly one valid JSON result or one error object
- **Executable evidence:** TestMiddlewareExecutesOnceAndReplaysResult
- **Executable evidence:** TestMiddlewareReplaysJSONRPCError
- **Executable evidence:** TestMiddlewareFailsClosedAtStorageAndPersistedReplayBoundaries
- **Fuzz evidence:** FuzzMalformedReplayFailsClosed
- **Public API:** idempotencyrpc.Response
- **Public API:** idempotencyrpc.Error
- **Public API:** idempotencyrpc.CallResult
- **Documentation:** docs/specification-decisions.md
- **Documentation:** docs/json-rpc.md
- **Additional authoritative source:** `{"id":"rfc8259-source","version":"RFC 8259","url":"https://www.rfc-editor.org/rfc/rfc8259.txt","specifications":["RFC 8259 JSON"]}`

## IDEMPOTENCY-DEC-008: JSON-RPC error codes and internal failure mapping

- **Status:** resolved
- **Owner:** idempotency maintainers
- **Classification:** optional behavior
- **Decision scope:** transport-specific
- **Specification:** JSON-RPC 2.0 Specification
- **Version:** 2.0 updated 2013-01-04
- **Source authority:** jsonrpc20-source
- **Section:** 5.1
- **Requirement strength:** not specified
- **Issue:** JSON-RPC reserves error ranges but leaves application errors open, and it does not define how an idempotency replay encoder failure should be represented.
- **Peer behavior:** JSON-RPC peers use -32603 for internal errors and permit application-defined codes outside reserved ranges.
- **Selected behavior:** Caller error codes are preserved; invalid or oversized handler output is durably mapped to -32603 internal error and is not rerun.
- **Rationale:** A handler that already ran must not be silently repeated because its response cannot be replayed.
- **Security consequences:** Invalid output becomes a bounded generic error without leaking payload details.
- **Resource consequences:** The internal-error envelope fits the configured minimum response limit.
- **Compatibility consequences:** Applications own valid use of reserved and application error ranges.
- **Wire consequences:** Terminal replay uses code -32603 and message internal error.
- **Upstream status:** The -32603 mapping uses the current JSON-RPC predefined error.
- **Reconsider when:** A versioned transport policy defines another terminal-error mapping.
- **Authoritative URL:** https://www.jsonrpc.org/specification
- **Credible interpretation:** Reject every application code
- **Credible interpretation:** Pass through caller error codes
- **Credible interpretation:** Rerun after invalid or oversized handler output
- **Credible interpretation:** Persist a terminal internal error
- **Executable evidence:** TestMiddlewareRecordsInvalidOrOversizedHandlerResponseAsTerminal
- **Executable evidence:** TestMiddlewareReplaysJSONRPCError
- **Public API:** idempotencyrpc.Error
- **Public API:** idempotencyrpc.Response
- **Public API:** idempotencyrpc.Middleware.Call
- **Documentation:** docs/specification-decisions.md
- **Documentation:** docs/json-rpc.md

## IDEMPOTENCY-DEC-009: JSON-RPC notification and batch ownership boundary

- **Status:** resolved
- **Owner:** idempotency maintainers
- **Classification:** omission
- **Decision scope:** normative
- **Specification:** JSON-RPC 2.0 Specification
- **Version:** 2.0 updated 2013-01-04
- **Source authority:** jsonrpc20-source
- **Section:** 4.1 and 6
- **Requirement strength:** MUST NOT
- **Issue:** JSON-RPC defines notification suppression and batch response assembly, while this middleware operates on one projected invocation without an id field.
- **Peer behavior:** Full JSON-RPC peers own request IDs, notification response suppression, batch scheduling, and response ordering.
- **Selected behavior:** The middleware processes one invocation and does not classify notifications, assemble batches, emit ids, or choose batch ordering; the surrounding transport owns those rules.
- **Rationale:** The projection lacks the fields required to implement those protocol contracts correctly.
- **Security consequences:** The middleware cannot accidentally emit a response to a notification on its own.
- **Resource consequences:** Batch width and transport parsing remain outside this package.
- **Compatibility consequences:** Callers must not present idempotencyrpc as a complete JSON-RPC server.
- **Wire consequences:** No full JSON-RPC request or response envelope is emitted.
- **Upstream status:** Notification and batch rules remain normative at the transport boundary.
- **Reconsider when:** A complete JSON-RPC transport adapter is added.
- **Authoritative URL:** https://www.jsonrpc.org/specification
- **Credible interpretation:** Infer notifications from missing data
- **Credible interpretation:** Assemble batches inside idempotency middleware
- **Credible interpretation:** Require the transport to classify notifications and decompose batches
- **Executable evidence:** TestMiddlewareValidatesInvocationIdentity
- **Executable evidence:** TestMiddlewareReturnsConflictAndInProgressWithoutExecuting
- **Public API:** idempotencyrpc.Request
- **Public API:** idempotencyrpc.Response
- **Public API:** idempotencyrpc.Middleware.Call
- **Documentation:** docs/specification-decisions.md
- **Documentation:** docs/json-rpc.md

## IDEMPOTENCY-DEC-010: HTTP Idempotency-Key syntax and application scoping

- **Status:** resolved
- **Owner:** idempotency maintainers
- **Classification:** interoperability policy
- **Decision scope:** transport-specific
- **Specification:** draft-ietf-httpapi-idempotency-key-header-07
- **Version:** draft-07 2025-10-15
- **Source authority:** idempotency-key-07-source
- **Section:** 2.1, 2.2, and 2.5
- **Requirement strength:** not specified
- **Issue:** The draft defines an Item Structured Header string, while the package historically accepts one trimmed header value and delegates semantic validation and scoping.
- **Peer behavior:** Draft-conforming HTTP peers serialize an sf-string; deployed provider profiles vary in accepted key syntax and scope.
- **Selected behavior:** Middleware requires one nonempty trimmed Idempotency-Key value and delegates syntax plus namespace, tenant, operation, caller, and value scoping to Key.
- **Rationale:** The stable v1 API predates draft-07 syntax and application identity cannot be inferred from the header alone.
- **Security consequences:** Applications must reject ambiguous or inappropriately scoped values in Key.
- **Resource consequences:** Constructed key parts are bounded by the semantic core.
- **Compatibility consequences:** The package does not claim full draft-07 structured-field conformance.
- **Wire consequences:** Header quotes and structured-field escaping are not decoded by middleware.
- **Upstream status:** The source is an expiring Internet-Draft and its Datatracker publication is monitored.
- **Reconsider when:** The draft becomes an RFC or a new major version adopts structured-field parsing.
- **Authoritative URL:** https://www.ietf.org/archive/id/draft-ietf-httpapi-idempotency-key-header-07.txt
- **Credible interpretation:** Parse RFC 8941 in middleware
- **Credible interpretation:** Accept any nonempty trimmed value
- **Credible interpretation:** Require the application Key callback to construct a bounded five-part identity
- **Executable evidence:** TestMiddlewareRequiresIdempotencyKey
- **Executable evidence:** TestMiddlewareRejectsKeyAndFingerprintFailures
- **Executable evidence:** TestNewKeyRequiresEveryIdentityPart
- **Public API:** idempotencyhttp.HeaderKey
- **Public API:** idempotencyhttp.KeyFunc
- **Public API:** NewKey
- **Documentation:** docs/specification-decisions.md
- **Documentation:** docs/http.md

## IDEMPOTENCY-DEC-011: HTTP fingerprint conflicts and status mapping

- **Status:** resolved
- **Owner:** idempotency maintainers
- **Classification:** interoperability policy
- **Decision scope:** transport-specific
- **Specification:** draft-ietf-httpapi-idempotency-key-header-07
- **Version:** draft-07 2025-10-15
- **Source authority:** idempotency-key-07-source
- **Section:** 2.4, 2.6, and 2.7
- **Requirement strength:** SHOULD
- **Issue:** The draft recommends fingerprints and distinguishes concurrent 409 from different-payload 422, while the stable middleware exposes one conflict family.
- **Peer behavior:** Draft-07 recommends 409 for concurrent work and 422 for key reuse with another payload.
- **Selected behavior:** Missing or invalid identity returns 400, unavailable storage returns 503, and both in-progress and fingerprint conflict return 409 with distinct Idempotency-Outcome values.
- **Rationale:** Stable machine-readable outcomes preserve the distinction even though v1 shares one HTTP status.
- **Security consequences:** Storage failure cannot authorize untracked handler execution.
- **Resource consequences:** Fingerprint generation remains application-bounded.
- **Compatibility consequences:** The 409 conflict mapping deliberately differs from draft-07's recommended 422 payload-conflict status.
- **Wire consequences:** Every response carries Idempotency-Outcome and replay adds Idempotency-Replayed true.
- **Upstream status:** Draft-07 status guidance is recommendation strength and remains work in progress.
- **Reconsider when:** A major transport profile adopts distinct 409 and 422 mappings.
- **Authoritative URL:** https://www.ietf.org/archive/id/draft-ietf-httpapi-idempotency-key-header-07.txt
- **Credible interpretation:** Use 422 for fingerprint mismatch
- **Credible interpretation:** Use 409 for every ownership conflict
- **Credible interpretation:** Expose package outcomes and let a higher transport profile remap them
- **Executable evidence:** TestMiddlewareReturnsExplicitProtocolOutcomes
- **Executable evidence:** TestMiddlewareRejectsKeyAndFingerprintFailures
- **Executable evidence:** TestBeginFailsClosedWhenStorageIsUnavailable
- **Public API:** idempotencyhttp.HeaderOutcome
- **Public API:** idempotencyhttp.HeaderReplayed
- **Public API:** idempotencyhttp.Middleware.Handler
- **Documentation:** docs/specification-decisions.md
- **Documentation:** docs/http.md
- **Additional authoritative source:** `{"id":"rfc9110-source","version":"RFC 9110","url":"https://www.rfc-editor.org/rfc/rfc9110.txt","specifications":["RFC 9110 HTTP Semantics"]}`

## IDEMPOTENCY-DEC-012: HTTP response replay and terminal failure policy

- **Status:** resolved
- **Owner:** idempotency maintainers
- **Classification:** optional behavior
- **Decision scope:** transport-specific
- **Specification:** RFC 9110 HTTP Semantics
- **Version:** RFC 9110
- **Source authority:** rfc9110-source
- **Section:** 6, 9.2.2, and 15
- **Requirement strength:** not specified
- **Issue:** HTTP and the Idempotency-Key draft do not define which response metadata is safe to persist or what to do after a handler produces an unreplayable response.
- **Peer behavior:** Provider idempotency profiles differ on cached failures, header replay, retention, and response size.
- **Selected behavior:** Middleware buffers one bounded response, persists only configured canonical header names, replays completed and terminal responses, and records terminal HTTP 500 after oversized output.
- **Rationale:** A handler that may have caused a side effect must not be rerun solely because replay encoding failed.
- **Security consequences:** Secret, hop-by-hop, and unbounded headers are excluded unless callers explicitly misconfigure ReplayHeaders.
- **Resource consequences:** Buffered and persisted response bytes have explicit maxima.
- **Compatibility consequences:** Streaming, flushing, hijacking, and full-duplex response interfaces are unsupported.
- **Wire consequences:** Replayed status, selected headers, and body match the stored bounded projection.
- **Upstream status:** This is a package transport profile constrained by RFC 9110, not an HTTP-wide requirement.
- **Reconsider when:** A named external response-replay profile is adopted.
- **Authoritative URL:** https://www.rfc-editor.org/rfc/rfc9110.txt
- **Credible interpretation:** Replay every response header
- **Credible interpretation:** Rerun after oversized output
- **Credible interpretation:** Buffer a bounded status, selected headers, and body and persist terminal 500 on encoding failure
- **Executable evidence:** TestMiddlewareExecutesOnceAndReplaysResponse
- **Executable evidence:** TestMiddlewareBoundsHandlerResponseAndRecordsTerminalFailure
- **Executable evidence:** TestMiddlewareDeduplicatesReplayHeaders
- **Fuzz evidence:** FuzzMalformedReplayFailsClosed
- **Public API:** idempotencyhttp.Options
- **Public API:** idempotencyhttp.ErrResponseTooLarge
- **Public API:** idempotencyhttp.Middleware.Handler
- **Documentation:** docs/specification-decisions.md
- **Documentation:** docs/http.md
- **Additional authoritative source:** `{"id":"idempotency-key-07-source","version":"draft-07 2025-10-15","url":"https://www.ietf.org/archive/id/draft-ietf-httpapi-idempotency-key-header-07.txt","specifications":["draft-ietf-httpapi-idempotency-key-header-07"]}`

## IDEMPOTENCY-DEC-013: Durable enforcement, expiry, and fencing semantics

- **Status:** resolved
- **Owner:** idempotency maintainers
- **Classification:** omission
- **Decision scope:** application-policy
- **Specification:** draft-ietf-httpapi-idempotency-key-header-07
- **Version:** draft-07 2025-10-15
- **Source authority:** idempotency-key-07-source
- **Section:** 2.3, 2.5, and 2.6
- **Requirement strength:** SHOULD
- **Issue:** The draft assigns key lifecycle and enforcement to the resource but does not define distributed ownership, crash ambiguity, takeover, or fencing.
- **Peer behavior:** HTTP idempotency implementations vary between caches, database uniqueness, durable state machines, and provider-specific replay stores.
- **Selected behavior:** Durable stores elect one current owner, preserve completed or terminal replay, distinguish fingerprint conflict, and require application fencing because lease expiry never proves old work stopped.
- **Rationale:** Exactly-once execution cannot be inferred across process death and unrelated side-effect stores.
- **Security consequences:** Unavailable durable ownership fails closed by default and stale owners cannot complete with old fences.
- **Resource consequences:** Leases, results, metadata, keys, retention, and cleanup are bounded.
- **Compatibility consequences:** Process-local memory is not a multi-replica durability claim.
- **Wire consequences:** Transport adapters project stable acquired, replayed, conflict, in-progress, terminal, and unavailable outcomes.
- **Upstream status:** Distributed durability and fencing are package-owned additions to draft enforcement guidance.
- **Reconsider when:** A standards profile defines equivalent crash, ownership, and fencing semantics.
- **Authoritative URL:** https://www.ietf.org/archive/id/draft-ietf-httpapi-idempotency-key-header-07.txt
- **Credible interpretation:** Treat a cache hit as exactly-once execution
- **Credible interpretation:** Let process-local leases authorize distributed side effects
- **Credible interpretation:** Use durable acquisition outcomes, bounded leases, monotonically increasing fences, and explicit unknown results
- **Executable evidence:** TestBeginMapsDurableOutcomesToExecutionDecisions
- **Executable evidence:** TestStoreConformance
- **Executable evidence:** TestBeginFailsClosedWhenStorageIsUnavailable
- **Public API:** Service
- **Public API:** Store
- **Public API:** Outcome
- **Public API:** Ownership
- **Public API:** FencingToken
- **Documentation:** docs/specification-decisions.md
- **Documentation:** docs/state-machine.md
- **Documentation:** docs/crash-semantics.md

## Authority review history

| Reviewed | Authority | Disposition | Decision impact |
| --- | --- | --- | --- |
| 2026-09-03 | RFC 9110 Erratum 9162 | Behavior-neutral. The proposed comma-space spelling concerns combining repeated HTTP field lines. The HTTP adapter preserves each configured response-header field value as a separate value and delegates wire serialization to Go's `net/http`; it does not define a combined-field-value syntax. | `IDEMPOTENCY-DEC-011` and `IDEMPOTENCY-DEC-012` remain unchanged, as do their conformance bindings. |

## Unresolved decisions

None for the currently supported v1 surfaces. New canonicalization policies,
transport projections, status mappings, digest framing, persistence semantics,
or external specification revisions require a new or superseding decision
before runtime behavior changes.
