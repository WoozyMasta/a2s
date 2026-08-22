// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

/*
Package a2s provides a client for Valve's Steam A2S server query protocol.

The package supports the primary A2S_INFO, A2S_PLAYER, and A2S_RULES
queries used by Source and GoldSource-compatible servers.
It also exposes methods for the obsolete A2S_SERVERQUERY_GETCHALLENGE
and A2A_PING wire requests for protocol compatibility.
Challenge exchanges, split responses, query deadlines,
and response validation are handled by the client.

Create a Client with New, NewWithString, or NewWithAddr.
Query methods accept a context.Context
and return typed results where the protocol has a defined structure.
GetRules preserves server rule order, duplicate names, and string values;
Rules.Map provides an explicitly lossy map conversion;
GetParsedRules and ParseRuleValues provide
optional heuristic conversion to convenient Go values.

Use [Client.GetInfo] for server metadata,
or [Client.GetInfoWithMeta] when query latency is also needed,
[Client.GetPlayers] for the current player list,
and [Client.GetRules] for server-defined key/value properties.
Use [Client.Query] when the complete logical response packet is needed.

A2S_INFO exposes the 16-bit AppID and optional EDF GameID separately.
Use [Info.EffectiveID] when an effective game identifier is needed.
Use [DecodeInfo] and [AppendInfo] for transport-independent A2S_INFO codecs.
Use [DecodePlayers] and [AppendPlayers] for standard A2S_PLAYER codecs;
The Ship's extended player response remains a separate API.
Use [DecodeRules] and [AppendRules] for ordered A2S_RULES codecs.

Complete query transactions on one Client are serialized.
Use separate clients when independent queries must run in parallel.
The client does not require an A2S_INFO request before querying players or rules.

Close the client when it is no longer needed.
Runnable API examples are available in the package example tests.

See the Valve protocol documentation for wire-level details: [Server queries].

[Server queries]: https://developer.valvesoftware.com/wiki/Server_queries
*/
package a2s
