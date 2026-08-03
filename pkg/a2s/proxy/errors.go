package proxy

import "errors"

var (
	// ErrCache identifies a cache operation or configuration failure.
	ErrCache = errors.New("a2s proxy: cache failure")

	// ErrCacheQuery identifies a query type that cannot be cached.
	ErrCacheQuery = errors.New("a2s proxy: unsupported cache query")

	// ErrCacheDisabled identifies a supported query that is not enabled in a cache.
	ErrCacheDisabled = errors.New("a2s proxy: cache query is disabled")

	// ErrCachePacket identifies a response packet that cannot answer its query.
	ErrCachePacket = errors.New("a2s proxy: invalid cache packet")

	// ErrPoller identifies a polling configuration or lifecycle failure.
	ErrPoller = errors.New("a2s proxy: poller failure")

	// ErrHandler identifies invalid proxy handler configuration or use.
	ErrHandler = errors.New("a2s proxy: handler failure")
)
