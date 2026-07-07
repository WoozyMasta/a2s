package server

import (
	"context"
	"net/netip"

	"github.com/woozymasta/a2s/pkg/a2s"
)

// PanicReport describes a panic recovered at the handler boundary.
type PanicReport struct {
	// Request is the request being processed when the panic occurred.
	Request *Request
	// Value is the value supplied to panic.
	Value any
	// Stack contains the recovered goroutine's stack trace.
	Stack []byte
}

// Request contains the remote endpoint and decoded query received by a server handler.
type Request struct {
	// Remote is the endpoint from which the query was received.
	Remote netip.AddrPort
	// Query is the decoded A2S request.
	Query a2s.Request
}

// Response is a logical response returned by a server handler.
//
// The concrete response types are intentionally limited
// to the response formats supported by this package.
// Use PacketResponse for an already decoded packet
// that should pass through without typed conversion.
type Response interface {
	isA2SResponse()
}

// InfoResponse contains a typed A2S_INFO response.
type InfoResponse struct {
	// Info is the decoded A2S_INFO response.
	Info a2s.Info
}

// PlayersResponse contains a typed A2S_PLAYER response.
type PlayersResponse struct {
	// Players contains the decoded player records in wire order.
	Players []a2s.Player
}

// RulesResponse contains a typed A2S_RULES response.
type RulesResponse struct {
	// Rules contains the decoded rules in wire order.
	Rules a2s.Rules
}

// PacketResponse contains a logical packet response that is written
// without converting it through one of the typed response formats.
type PacketResponse struct {
	// Packet is the decoded logical packet to write to the client.
	Packet a2s.Packet
}

// Handler processes one decoded A2S request and returns a logical response.
type Handler interface {
	Handle(context.Context, *Request) (Response, error)
}

// HandlerFunc adapts a function to the Handler interface.
type HandlerFunc func(context.Context, *Request) (Response, error)

// Handle calls f with the request context and decoded request.
func (f HandlerFunc) Handle(ctx context.Context, req *Request) (Response, error) {
	return f(ctx, req)
}

// PanicReporter receives recovered handler panics.
type PanicReporter func(context.Context, PanicReport)

// Middleware wraps a Handler with additional request or response behavior.
type Middleware func(Handler) Handler

func (InfoResponse) isA2SResponse()    {}
func (PlayersResponse) isA2SResponse() {}
func (RulesResponse) isA2SResponse()   {}
func (PacketResponse) isA2SResponse()  {}
