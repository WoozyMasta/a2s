package server

import "errors"

var (
	// ErrResponse identifies a failure while normalizing a handler response.
	ErrResponse = errors.New("a2s server: response normalization failed")

	// ErrResponseQuery identifies a response that cannot answer the request type.
	ErrResponseQuery = errors.New("a2s server: response does not match query")

	// ErrChallengeProvider identifies failure to initialize a challenge provider.
	ErrChallengeProvider = errors.New("a2s server: challenge provider initialization failed")

	// ErrDrop tells the server to discard the request without sending a response.
	// It is not a server failure and should not be reported
	// as one by the transport implementation.
	ErrDrop = errors.New("a2s server: drop response")
)
