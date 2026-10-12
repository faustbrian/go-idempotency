# Source documentation audit

## Acceptance and owned boundary

This Tier A pass owns comments and directly contradicted current prose at main
fe415c5f6bba8be80431901335a7bb8a5f7b251e. Runtime AST/tokens/literals/directives,
SQL/Lua, migrations, API/dependency/source pins, official fixtures and historical
authorities remain exact. Consumers are semantic ownership/replay, canonical
fingerprints, memory/PostgreSQL/Valkey stores, HTTP/RPC/command/queue/webhook/
outbox adapters, observability and the intentionally retained v1 ecosystem.
Named risks are inaccurate authentication/privacy, copy and cancellation claims,
retention/fencing scope, transaction commit ambiguity, callback/replay lifecycle
and resource admission. No runtime, provider, protocol or campaign repair.

## Reproducible before-write inventory

Git and source-docs inventory identify two modules, 94 Go files (38 source,
56 test), 23 directories and 140 structural gaps: 139 exported members/methods
and a canonical Valkey package comment. All root, legacy and canonical adapter
paths, conformance helpers, fixtures, eleven fuzz targets (nine configured),
benchmarks/Examples, live backend and intentional v1 compatibility tests are included. No tagged,
platform, generated/vendor Go or generator/embed directive is present. SQL/Lua
strings and their runtime substitution are executable artifacts, not prose.

Native .golib.yaml configures apidiff api/baseline.txt, not full GoDoc; preserve
that frozen API input. Render all operational root-module packages separately;
compatibility/ecosystem is a test-only module bound to released v1 imports, not
the current v2 source. No repin or composition campaign belongs in this pass.
Native offline specification checking owns the typed decision register; retain
source authority and append-only history unless a concrete contradiction needs
reconciliation. Make check/ci provision services and races; conformance installs
a locked JavaScript peer, so neither is a prose gate.

## Missing versus technical investigation

Missing comments concentrate in ownership record/request data, service results,
adapter callbacks/options and transactional ports. Trace fingerprint conflict
precedence, expiry versus purge, detached cleanup, unrenewed handler leases,
unknown commit outcomes, shared snapshots/context and trusted callback limits.
Investigate custom instrumentation (synchronous, panic-contained but unbounded),
key authentication supplied by callers, retention resetting fences, provider
clock/admission differences and response persistence versus transport delivery.
Existing security/backend/conformance evidence is not re-certified by this pass.

## Smallest sufficient proof

Pinned Go 1.27.2 formatting and all 94-file AST/raw-token/literal/directive
identity; useful native GoDoc renders and selected existing Examples where
present; structural source, local documentation and offline specification checks;
pinned full/changed/staged secrets and ordinary complete-diff inspection.
No generator, service, composition, active fuzz, race/leak, benchmark or online
conformance campaign. Own and immediately remove all three caches/helpers.
Coordinator owns remote main, exact-source CI and delivery.

## Source-derived coverage and disposition

| Boundary | Clarified source contract |
| --- | --- |
| Core identity/state/service | Caller-established trust; byte rather than Unicode admission; conflict before replay; copied snapshots; exclusive live-owner proof; tracked success is not persistence qualification. |
| Canonical fingerprints | Owned RFC 8785 policy and deliberate negative-zero rejection; input/depth checks before the pinned transformer, output cap afterward, exposed error causes, no acquisition/heap sandbox. |
| Memory | Process-local retained generations, fixed capacity, synchronous non-reentrant clock/token calls under one mutex, cancellation checked before locking, no retention worker. |
| PostgreSQL | Borrowed pool and caller migrations; absent-key serialization, server timestamps, provisional caller transactions, possibly populated results before commit error, detached rollback without a new deadline, purge resets the fencing domain. |
| Valkey | Borrowed client and startup-only version/eviction checks; one-key Lua/server time, whole-millisecond durations, token callback even for replay, TTL resets counters, infrastructure-owned failover durability. |
| HTTP/RPC | Trusted synchronous policy callbacks, no automatic renewal/handler termination; HTTP selects the first key-header value and replays stored global-budget projections; RPC protocol errors complete while invalid/oversized responses fail terminally; fresh RPC response aliases remain possible. |
| Command/queue/webhook/outbox | Fenced attempt rather than exactly-once effect; detached failed-handler cleanup; broker/provider settlement and authentication remain caller-owned; transaction helper inserts then conditionally completes without commit/rollback ownership. |
| Instrumentation and fixtures | Trusted unbounded custom reason/correlation text, synchronous latency and panic-contained attempted delivery; owned HMAC secret copying/lifetime; deterministic test tokens/clocks, live-provider benchmark shapes, codec-only versus server fuzz boundaries and v1-only compatibility module. |

The 13 current resolved specification decisions remain consistent with these
source policies; their typed records, history, authorities and official inputs
are unchanged. HTTP's first-value selection is documented explicitly without
inventing a duplicate-header rejection rule. Deprecation support horizons and
legacy public type identity are retained. Unknown business decisions, backend
configuration and downstream fence enforcement are application responsibilities,
not invented guarantees or source repairs.

## Executed proof and boundaries

Go 1.27.2 formatting, all 94-file exact AST/raw-token/literal/implicit-semicolon/
directive identity and comment HTML rendering passed in an owned session.
All 22 operational root-module GoDocs rendered successfully. The existing
ExampleService_Begin passed with `go test ./... -run '^Example' -count=1
-timeout=2m`; the other 21 packages compiled without selecting runtime tests.
The configured apidiff baseline, both modules/dependency inputs, SQL/Lua,
authorities and all 22 Markdown fenced artifacts remain exact. Structural
source checking fell from 140 findings to zero; direct changed local links and
native offline `golib specification check` passed for 13 decisions. Full baseline,
current and changed snapshots passed pinned Gitleaks 8.30.1 with ignore annotations
disabled and complete redaction. Final 94-file identity/HTML refresh and seven
affected GoDoc renders passed;
other unchanged package renders and Example/provider inputs are reused, not
rerun. The final Record encryption/redaction ownership refinement also rendered
in root GoDoc. A redundant comment-stripped source comparison initially rejected
gofmt field padding; the corrected nonwhitespace comparison passed. This was
not an executable source mismatch, and its owned resources were removed.

An earlier `golib docs check` was mistakenly treated as structure-only and
launched before the explicit Go slot without task-owned cache overrides. It
also compiled the intentional compatibility module with no tests selected.
This was a cache/serialization boundary violation, not v1 runtime or current-v2
composition proof. No unexpected worktree artifacts remain; shared user caches
were not removed. Subsequent Go sessions own and remove all three caches and
helpers. No service, compatibility runtime, race/leak, active fuzz, benchmark,
JavaScript differential or release campaign was executed by the bounded pass.
Coordinator retains exact-source remote CI and delivery ownership.
