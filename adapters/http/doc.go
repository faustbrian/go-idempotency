// Package idempotencyhttp provides buffered net/http middleware backed by an
// idempotency.Service. Applications must supply semantic key and canonical
// fingerprint policies; the package does not infer business identity from
// unstable transport details. Use New/Handler behind caller authentication and
// request-body/deadline admission. Handler leases are not renewed automatically;
// external effects and replay persistence are separate unless application code
// combines them. This is buffered response replay, not streaming/Hijacker/Flusher.
package idempotencyhttp
