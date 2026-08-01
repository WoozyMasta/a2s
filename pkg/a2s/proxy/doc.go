// Package proxy provides reusable building blocks for a cached A2S proxy.
//
// The package operates on logical a2s.Packet values rather than UDP datagrams.
// Transport, challenge validation, packetization,
// and listener lifecycle remain responsibilities of pkg/a2s/server.
package proxy
