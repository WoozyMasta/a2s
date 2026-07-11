/*
Package server defines transport-independent contracts for an A2S server.

The package contains request, handler, middleware, response, and UDP runtime types.
Server workers decode requests, apply challenge handling,
dispatch the handler, encode responses, and packetize Source responses.

State provides an atomic pre-encoded handler
for static or periodically updated INFO, PLAYER, and RULES responses.

Use NormalizeResponse to turn a handler response into a logical A2S packet
when implementing a custom transport or packetizer.

Use SecureChallengePolicy by default for Internet-facing servers;
LegacyChallengePolicy and NoChallengePolicy are explicit alternatives.
NewChallengeGate applies a policy and provider before a handler receives a request.
*/
package server
