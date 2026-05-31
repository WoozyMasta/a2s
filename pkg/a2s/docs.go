/*
Package a2s provides a client for Valve's Steam A2S server query protocol.

The package supports the primary A2S_INFO, A2S_PLAYER, and A2S_RULES
queries used by Source and GoldSource-compatible servers.
It also exposes the deprecated A2S_SERVERQUERY_GETCHALLENGE
and A2A_PING methods for protocol compatibility.
Challenge exchanges, split responses, query deadlines,
and response validation are handled by the client.

Create a Client with New, NewWithString, or NewWithAddr.
Query methods accept a context.Context
and return typed results where the protocol has a defined structure.
GetRules preserves the server's rule values as strings;
GetParsedRules and ParseRuleValues provide
optional heuristic conversion to convenient Go values.

Use [Client.GetInfo] for server metadata,
[Client.GetPlayers] for the current player list,
and [Client.GetRules] for server-defined key/value properties.

Complete query transactions on one Client are serialized.
Use separate clients when independent queries must run in parallel.
The client does not require an A2S_INFO request before querying players or rules.

Close the client when it is no longer needed.
Runnable API examples are available in the package example tests.

See the Valve protocol documentation for wire-level details: [Server queries].

[Server queries]: https://developer.valvesoftware.com/wiki/Server_queries
*/
package a2s
