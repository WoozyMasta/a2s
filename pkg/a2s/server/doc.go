/*
Package server defines transport-independent contracts for an A2S server.

The package contains request, handler, middleware, response, and UDP runtime types.
Server workers decode requests, apply challenge handling,
dispatch the handler, encode responses, and packetize configured response frames.

State provides an atomic pre-encoded handler
for static or periodically updated INFO, PLAYER, and RULES responses.

Use NormalizeResponse to turn a handler response into a logical A2S packet
when implementing a custom transport or packetizer.

Use SecureChallengePolicy by default for Internet-facing servers;
LegacyChallengePolicy and NoChallengePolicy are explicit alternatives.
NewChallengeGate applies a policy and provider before a handler receives a request.

# Security model

New enables SecureChallengePolicy by default.
INFO, PLAYER, and RULES requests must first present
a valid challenge bound to the remote UDP endpoint.
The default stateless provider uses a random process secret
and a short time window; it does not store per-client state.

LegacyChallengePolicy leaves INFO unchallenged for compatibility with older clients and servers.
Use it only when that compatibility is required:
an untrusted sender can otherwise cause the server to send INFO responses
without first proving reachability at the advertised endpoint.
NoChallengePolicy disables this protection and is intended for local tests or a trusted network.

Challenges are an anti-amplification and request-validation mechanism, not authentication.
A2S is UDP-based, so source addresses can be spoofed
and the protocol provides neither confidentiality nor general message integrity.
Put an Internet-facing deployment behind the network controls appropriate for its environment
and add application-specific access control in a Handler or Middleware when needed.

# Operational model

Serve uses a fixed worker pool.
DefaultWorkerCount derives the worker count from GOMAXPROCS;
WithWorkers can set an explicit count.
Handler calls may run concurrently and must protect their own mutable state.
WithMaxRequestSize limits accepted request datagrams,
and the packetizers enforce logical response and fragment limits.
Oversized or malformed requests and handler responses are dropped by the transport loop.

Handler panics are recovered so one request cannot terminate the serving loop.
Use WithPanicReporter when operational reporting is required.
Shutdown cancels active handler contexts and waits for workers.
Serve does not close a caller- owned net.PacketConn;
ListenAndServe owns and closes the socket it creates.

The package does not provide authentication, rate limiting, caching, or a built-in upstream proxy.
Compose those behaviors with Handler and Middleware.
Use PacketResponse for a logical packet that must pass through without typed decode/re-encode;
the server still applies its configured challenge gate and packetization limits.
*/
package server
