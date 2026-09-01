.PHONY: conformance

conformance:
	go test ./canonical ./idempotencyhttp ./idempotencyrpc . -run '^(TestJSONMatchesPinnedRFC8785Fixtures|TestJSONUsesRFC8785CanonicalForm|TestJSONRejectsHostileOrAmbiguousInput|TestMiddlewareRequiresIdempotencyKey|TestMiddlewareReturnsExplicitProtocolOutcomes|TestMiddlewareExecutesOnceAndReplaysResult|TestMiddlewareReplaysJSONRPCError|TestMiddlewareRecordsInvalidOrOversizedHandlerResponseAsTerminal|TestFingerprintUsesAnExplicitVersion|TestNewHMACKeyHasherProtectsLogicalIdentity)$$' -count=1
	npm ci --ignore-scripts --no-audit --no-fund
	node ./scripts/check-jcs-differential.mjs
