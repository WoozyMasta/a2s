# Arma3 Server Browser v3 and DayZ Server Browser v2 Protocols

Both protocols are designed for server information transmission
and share a similar structure, differing only in minor details.
They are implemented as nested sub-protocols
within the Steam `A2S_RULES` protocol.
The keys encode chunk metadata (current index and total count),
while the values store the actual chunks.

Server details are transmitted through the `A2S_RULES` protocol
in messages formatted as described below.

During decoding,
the parser replaces these escape pairs with their original bytes:

* `{0x01, 0x01}` → `0x01`
* `{0x01, 0x02}` → `0x00`
* `{0x01, 0x03}` → `0xFF`

During encoding, [AppendEscapeSequences] applies the inverse mapping:

* `0x01` → `{0x01, 0x01}`
* `0x00` → `{0x01, 0x02}`
* `0xFF` → `{0x01, 0x03}`

The later paging stage splits the escaped message into fragments
and places them into `A2S_RULES` key-value pairs.

If the entire message exceeds `1400` bytes
(the maximum size of a single UDP packet in the Steam API),
some data must be truncated.
However, in practice, a buffer of up to `8192` bytes is often sufficient,
allowing servers to respond with all data in a single packet.

Background information:
<https://community.bistudio.com/wiki/Arma_3:_ServerBrowserProtocol3>

## Arma3 Server Browser v3 Protocol

![Arma3](a3sb.png)

## DayZ Server Browser v2 Protocol

![DayZ](dzsb.png)

[AppendEscapeSequences]: https://pkg.go.dev/github.com/WoozyMasta/a2s/pkg/a3sb#AppendEscapeSequences
