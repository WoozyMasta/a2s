<!-- omit in toc -->
# A2S

<!-- markdownlint-disable-next-line MD033 -->
<img src="winres/icon64.png" alt="MetricZ" align="left" width="64">

Powerful command-line utility and Go packages for querying Steam A2S
server information.
Built with specific support for Arma 3 and DayZ, this tool provides a
seamless way to retrieve essential server data.

## Description

Features

A2S supports querying Steam servers using the following methods:

* `A2S_INFO`: Retrieve basic information about the server, such as name,
  map, and player count. _(CLI and package)_
* `A2S_PLAYER`: Get details about each player currently on the server.
  _(CLI and package)_
* `A2S_RULES`: Fetch server-specific rules and settings. _(CLI and package)_
* `A2S_SERVERQUERY_GETCHALLENGE`: Request a challenge number for use in
  player and rules queries. _(only package)_
* `A2A_PING`: Measure the ping time to the server for latency insights.
  _(only package)_

Additionally, this tool features an extension for the Arma 3 Server Browser
Protocol (A3SBP). This extension overrides the standard `GetRules()` method,
enabling compatibility with the unique protocol used by Arma 3's server
browser.

## CLI Installation

You can download the latest version of the programme by following the links:

|           | MacOS              | Linux             | Windows             |
| --------- | ------------------ | ----------------- | ------------------- |
| **AMD64** | [a2s-darwin-amd64] | [a2s-linux-amd64] | [a2s-windows-amd64] |
| **ARM64** | [a2s-darwin-arm64] | [a2s-linux-arm64] | [a2s-windows-arm64] |

You can also use the command (for Linux amd64, adjust for your platform):

```bash
curl -#SfLo /usr/bin/a2s \
  https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-linux-amd64
chmod +x /usr/bin/a2s
a2s -h && a2s -v
```

## A2S CLI

Command-line utility for querying Steam A2S server information with support for Arma 3 and DayZ servers.

The utility supports the following commands:
* `info` - Retrieve server information `A2S_INFO`
* `rules` - Retrieve server rules `A2S_RULES`
* `players` - Retrieve player list `A2S_PLAYERS`
* `all` - Retrieve all available server information
* `ping` - Ping the server with `A2S_INFO`

For detailed information about available options and flags, run `a2s --help`.

## Package

### A2S

Example of use:

```go
client, err := a2s.New("127.0.0.1", 27016)
if err != nil {
  panic(err)
}
defer client.Close()

client.SetBufferSize(2048)
client.SetDeadlineTimeout(3)

info, err := client.GetInfo()
if err != nil {
  panic(err)
}

rules, err := client.GetRules()
if err != nil {
  panic(err)
}
```

* `github.com/woozymasta/a2s.GetInfo()` -> `A2S_INFO`
* `github.com/woozymasta/a2s.GetPlayers()` -> `A2S_PLAYER`
* `github.com/woozymasta/a2s.GetRules()` -> `A2S_RULES`
* `github.com/woozymasta/a2s.GetChallenge()` -> `A2S_SERVERQUERY_GETCHALLENGE`
* `github.com/woozymasta/a2s.GetPing()` -> `A2A_PING`

### A3SB

Example of use:

```go
client, err := a2s.New("127.0.0.1", 27016)
if err != nil {
  panic(err)
}
defer client.Close()

// Wrap client
a3Client := &a3sb.Client{Client: client}

// Game id must be passed as the second argument to properly read the Arma 3 or Dayz rules
rules, err := a3Client.GetRules(221100)
if err != nil {
  panic(err)
}

// Can also perform standard a2s methods
info, err := a3Client.GetInfo()
if err != nil {
  panic(err)
}
```

## Protocol Documentation

For a deeper understanding of the protocols used, refer to the official documentation:

* [Steam Server Queries][]
* [Arma 3 Server Browser Protocol v3][]
* [A3SB Protocol v3 Specification 🇬🇧][]
* [A3SB Protocol v3 Specification 🇷🇺][]

## Tested Games

During development, the functionality of `a2s` was
thoroughly tested across a diverse range of popular games that utilize the
Steam server query protocols.
This ensures compatibility and reliable performance.

The following games were used for testing:

* Counter-Strike 1.6
* Counter-Strike: Source
* Team Fortress 2
* Project Zomboid
* Valheim
* Rust
* Garry's Mod
* Insurgency
* Arma 3
* DayZ
* 7 Days to Die
* ARK: Survival Evolved
* Conan Exiles
* Unturned

These games represent a wide array of server types and implementations,
highlighting the versatility of the tools in querying servers effectively
across various scenarios.

Feel free to contribute or report issues if you find any compatibility
problems with additional games!

> [!NOTE]  
> Implementation of `bzip2` compression for multi-packet response is not
> implemented, as I did not find any servers that would respond in this
> format. If you know of one, please write me in issue.

## 👉 [Support Me](https://gist.github.com/WoozyMasta/7b0cabb538236b7307002c1fbc2d94ea)

Your support is greatly appreciated!

<!-- Links -->
[logo-a2s]: assets/a2s.png

[Steam Server Queries]: https://developer.valvesoftware.com/wiki/Server_queries
[Arma 3 Server Browser Protocol v3]: https://community.bistudio.com/wiki/Arma_3:_ServerBrowserProtocol3
[A3SB Protocol v3 Specification 🇬🇧]: https://github.com/WoozyMasta/a2s/blob/master/pkg/a3sb/docs/README.md "🇬🇧"
[A3SB Protocol v3 Specification 🇷🇺]: https://github.com/WoozyMasta/a2s/blob/master/pkg/a3sb/docs/README_ru.md "🇷🇺"

[a2s-darwin-arm64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-darwin-arm64 "MacOS arm64 file"
[a2s-darwin-amd64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-darwin-amd64 "MacOS amd64 file"
[a2s-linux-amd64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-linux-amd64 "Linux amd64 file"
[a2s-linux-arm64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-linux-arm64 "Linux arm64 file"
[a2s-windows-amd64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-windows-amd64.exe "Windows amd64 file"
[a2s-windows-arm64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-windows-arm64.exe "Windows arm64 file"
