# Upstream authority review history

This append-only record preserves reviewed changes to the authorities monitored
by [`monitoring.json`](monitoring.json). A monitoring digest changes only after
the corresponding upstream delta has been classified against the applicable
specification decisions.

## 2026-09-09: RFC 9110 errata

- **Authority:** `rfc9110-errata`
- **URL:** https://errata.rfc-editor.org/rfc9110
- **Previous SHA-256:**
  `1f6790054c0cdb2f2a70a94fa2b9c73b09a4ee0578a32b4a3006ed0ecfaac86d`
- **Reviewed SHA-256:**
  `cec32fd170146656d933f627b512f2e027ae5c3592f5ec7760c3627493b30505`
- **Retrieved and reviewed:** 2026-09-09
- **Applicability:** `IDEMPOTENCY-DEC-011` and `IDEMPOTENCY-DEC-012`
- **Disposition:** Behavior-neutral because the package does not parse the
  collected RFC 9110 ABNF or define HTTP wire serialization.

[Errata ID 9164](https://errata.rfc-editor.org/eid9164/) was reported on
2026-09-07 against RFC 9110 Appendix A. It documents semantically equivalent
normalizations between the collected ABNF and the rules in the body. The
idempotency HTTP adapter uses Go's `net/http`, persists configured response
header values separately, and does not parse or regenerate the RFC grammar.
The immutable RFC 9110 text remains byte-identical to the recorded source
digest. No selected behavior, decision, conformance binding, or executable
evidence changes. Reconsider this disposition if Errata ID 9164 is verified
with a semantic grammar change or the package begins parsing HTTP ABNF.

## 2026-09-03: RFC 9110 errata

- **Authority:** `rfc9110-errata`
- **URL:** https://errata.rfc-editor.org/rfc9110
- **Previous SHA-256:**
  `38bd006c96f8963d58573f704c5313a5f81968b90738c03ade0b036ec7bbdf4b`
- **Reviewed SHA-256:**
  `1f6790054c0cdb2f2a70a94fa2b9c73b09a4ee0578a32b4a3006ed0ecfaac86d`
- **Retrieved and reviewed:** 2026-09-03
- **Applicability:** `IDEMPOTENCY-DEC-011` and `IDEMPOTENCY-DEC-012`
- **Disposition:** Behavior-neutral because repeated response-header values
  remain separate and wire serialization belongs to Go's `net/http`.

[Errata ID 9162](https://errata.rfc-editor.org/eid9162/) was reported on
2026-09-01 against RFC 9110 Section 5.2. It proposes comma-space rather than
comma when repeated field-line values are combined. The adapter preserves each
configured response-header value separately and delegates wire serialization
to Go's `net/http`; it does not define combined-field-value syntax. The erratum
remains Reported, and current source and tests confirm that no selected
behavior, decision, or conformance binding changes. Reconsider this disposition
if Errata ID 9162 becomes Verified or the adapter begins combining repeated
field values itself.
