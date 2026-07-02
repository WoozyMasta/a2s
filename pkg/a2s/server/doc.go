/*
Package server defines transport-independent contracts for an A2S server.

The package contains request, handler, middleware, and response types only.
A UDP listener and protocol dispatch are implemented separately.

Use NormalizeResponse to turn a handler response into a logical A2S packet
when implementing a custom transport or packetizer.

Use SecureChallengePolicy by default for Internet-facing servers;
LegacyChallengePolicy and NoChallengePolicy are explicit alternatives.
*/
package server
