// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

/*
Package a3sb parses Arma 3 Server Browser Protocol responses carried by A2S_RULES.

Wrap an *a2s.Client in [Client] and use [Client.GetRules]
to retrieve typed Arma 3 or DayZ data.

A known non-zero game AppID selects its explicit layout;
use constants from github.com/woozymasta/a2s/pkg/appid,
such as appid.Arma3 or appid.DayZ.
The [Client.GetRulesArma3] and [Client.GetRulesDayZ] helpers
select the corresponding layouts directly.

The parsed [Rules.Layout] records the selected binary layout.
The [Rules.GetAppID] result remains the requested or inferred game identity.
Use [AppendBinary] for the inner binary payload;
it does not perform A3SB escaping or A2S_RULES page generation.
Use [AppendEscapeSequences] before [EncodePages] to build ordered A2S_RULES page entries;
a page size of zero selects [DefaultPageSize].

Pass game == 0 when the server's game is unknown.
The response is classified from the A2S_RULES payload
without an additional A2S_INFO request.
Native A2S rules are returned in [Rules.ExtraRules] with [Rules.Version] == 0;
recognized A3SB responses are parsed into typed fields with a non-zero [Rules.Version].

For A3SB responses, [Rules.ExtraRules] contains ordinary outer A2S properties
that are not represented by typed fields.
Binary A3SB page carriers are consumed by the parser
and are not exposed as ordinary rule strings.
The parsed [Rules] model is intentionally lossy;
use the original response for byte-preserving round-trips or transparent proxying.

Close the embedded A2S client when it is no longer needed.
Runnable API examples are available in the package example tests.

See the protocol documentation for Arma 3: [Protocol v3] and [Protocol v2].
Additional implementation notes are available in [A3SB protocol notes].

[Protocol v3]: https://community.bistudio.com/wiki/Arma_3:_ServerBrowserProtocol3
[Protocol v2]: https://community.bistudio.com/wiki/Arma_3:_ServerBrowserProtocol2
[A3SB protocol notes]: https://github.com/WoozyMasta/a2s/blob/master/pkg/a3sb/docs/README.md
*/
package a3sb
