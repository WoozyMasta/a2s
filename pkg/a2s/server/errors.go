package server

import "errors"

// ErrDrop tells the server to discard the request without sending a response.
// It is not a server failure and should not be reported
// as one by the transport implementation.
var ErrDrop = errors.New("a2s server: drop response")
