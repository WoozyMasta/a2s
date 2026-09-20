# A2S

<!-- markdownlint-disable-next-line MD033 -->
<img src="winres/icon256.png" alt="A2S Logo" align="right" width="128">

Powerful command-line utility and Go toolkit
for working with Steam A2S game servers.
Query server information, players, and rules;
build A2S servers and cached UDP proxies;
and handle Arma 3 and DayZ Server Browser Protocol responses.

<!-- markdownlint-disable-next-line MD033 -->
## Supported protocols and components <br clear="right"/>

### Supported protocols

* Source and GoldSource Server Queries:
  `A2S_INFO`, `A2S_PLAYER`, and `A2S_RULES`.
* Source and GoldSource split responses, including reordered fragments;
  Source bzip2-compressed payloads.
* A2S challenge exchanges and the legacy-compatible
  `A2S_SERVERQUERY_GETCHALLENGE` and `A2A_PING` requests exposed by the Go API.
* Arma 3 and DayZ Server Browser Protocol ([A3SB][]) responses
  carried by `A2S_RULES`, with automatic detection,
  explicit layouts and native A2S fallback.

### Project components

* `cmd/a2s`: CLI for queries, diagnostics and cached proxy operation;
* `pkg/a2s`: context-aware A2S client and transport-independent codecs;
* `pkg/a2s/server`: UDP A2S server runtime with challenges and packetization;
* `pkg/a2s/proxy`: reusable cache, polling, relay, and rate-limit components;
* `pkg/a3sb`: typed Arma 3 and DayZ Server Browser Protocol parser and codec;
* `pkg/keywords`: typed parsers for Arma 3 and DayZ `A2S_INFO` keywords;
* `pkg/appid`: curated Steam AppID registry for A2S-compatible games.

## Installation

### Go library

```shell
go get github.com/woozymasta/a2s
```

### CLI

Download the latest CLI release for your platform:

Arch/OS | macOS | Linux | Windows
------- | ----- | ----- | -------
**AMD64** | [a2s-darwin-amd64][] | [a2s-linux-amd64][] | [a2s-windows-amd64][]
**ARM64** | [a2s-darwin-arm64][] | [a2s-linux-arm64][] | [a2s-windows-arm64][]

Download the latest release with curl in Bash:

```shell
case "$(uname -s)" in
  Linux*) OS=linux;;
  Darwin*) OS=darwin;;
  MINGW*|MSYS*|CYGWIN*) OS=windows; EXT=.exe;;
  *) exit 1;;
esac
case "$(uname -m)" in
  x86_64|amd64) ARCH=amd64;;
  aarch64|arm64) ARCH=arm64;;
  *) exit 1;;
esac
BIN="./a2s$EXT"

curl -#SfLo "$BIN" \
  "https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-$OS-$ARCH$EXT"
chmod +x "$BIN"
"$BIN" -h && "$BIN" -v
```

### Container image

Prebuilt images are available from GitHub Container Registry and Docker Hub:

* `ghcr.io/woozymasta/a2s:latest`
* `docker.io/woozymasta/a2s:latest`

Example using a container image:

```shell
docker pull ghcr.io/woozymasta/a2s:latest
docker run --rm -ti \
  -e A2S_TIMEOUT=3s \
  -e A2S_BUFFER_SIZE=8192 \
  ghcr.io/woozymasta/a2s:latest info host:port
```

## CLI usage

The CLI queries A2S servers, formats responses, and runs a cached proxy.
See the [CLI reference][CLI] for all commands, options, defaults,
environment variables, and generated examples.

### Commands

* `info`: query server metadata;
* `players`: list current players;
* `rules`: query native A2S or A3SB rules;
* `all`: query metadata, rules, and players together;
* `ping`: measure repeated query response times;
* `proxy`: expose a cached UDP endpoint for an upstream server.

### Examples

```shell
# Query server metadata.
a2s info 127.0.0.1:27015

# Export all responses as one JSON document.
a2s all 127.0.0.1:27015 --format json | jq

# Parse Arma 3 server-browser rules explicitly.
a2s rules 127.0.0.1:2303 --game arma3

# Print only ping values for five queries.
a2s ping 127.0.0.1:27015 --ping-count 5 --compact

# Cache an upstream server on a local UDP endpoint.
a2s proxy 127.0.0.1:27015 --listen :27016
```

## Module usage

The Go packages cover both sides of the protocol:
querying existing servers and serving or proxying A2S responses.
These examples show the main building blocks without covering every option.

### Client example

Use `pkg/a2s` for standard server queries.

```go
ctx := context.Background()
client, err := a2s.New("127.0.0.1", 27015, a2s.WithTimeout(3*time.Second))
if err != nil {
  panic(err)
}
defer client.Close()

info, err := client.GetInfo(ctx)
if err != nil {
  panic(err)
}
_ = info
```

#### A3SB client

Wrap an existing `a2s.Client` with `pkg/a3sb` for Arma 3 and DayZ rules:

```go
a3sClient := &a3sb.Client{Client: client}
rules, err := a3sClient.GetRules(ctx, 0) // Automatic classification.
if err != nil {
  panic(err)
}
_ = rules

// Explicit layout: a3sClient.GetRules(ctx, appid.DayZ)
```

### Server example

To publish your own server data,
implement a `Handler` and pass it to `server.New`.
The server handles UDP transport, challenge validation,
and response packetization.

```go
handler := server.HandlerFunc(func(
  _ context.Context,
  request *server.Request,
) (server.Response, error) {
  if request.Query.Type != a2s.InfoRequest {
    return nil, server.ErrDrop
  }
  return server.InfoResponse{Info: a2s.Info{
    Format: a2s.InfoFormat(a2s.ResponseInfo),
    Name:   "Example A2S server",
  }}, nil
})

srv, err := server.New(handler)
if err != nil {
  panic(err)
}
conn, err := net.ListenPacket("udp", ":27015")
if err != nil {
  panic(err)
}
defer conn.Close()
go srv.Serve(conn)
defer srv.Shutdown(context.Background())
```

### Proxy example

Compose `pkg/a2s/proxy` with `pkg/a2s/server` to serve cached responses and
relay uncached queries. Use a `Poller` to refresh selected cache entries from
the upstream client.

```go
cache, _ := proxy.NewCache([]a2s.QueryType{
  a2s.InfoRequest, a2s.RulesRequest,
})
handler, _ := proxy.NewHandler(cache, upstream, proxy.HandlerConfig{
  ChallengeProvider: provider,
  LocalPing:         true,
})
srv, _ := server.New(handler, server.WithChallengeProvider(provider))
```

Run a `Poller` to refresh the cache and serve `srv` on a UDP `PacketConn`.

## Protocol Documentation

For a deeper understanding of the protocols used,
refer to the official documentation:

* [Steam Server Queries][]
* [Arma 3 Server Browser Protocol v3][]
* [A3SB Protocol v3 Specification][A3SB]

## Tested Games

During development, the functionality of `a2s` was thoroughly tested
across a diverse range of popular games
that utilize the Steam server query protocols.

The implementation has been tested against servers from these games:

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

## 👉 [Support Me][]

Your support is greatly appreciated!

<!-- Links -->
[Steam Server Queries]: https://developer.valvesoftware.com/wiki/Server_queries
[Arma 3 Server Browser Protocol v3]: https://community.bistudio.com/wiki/Arma_3:_ServerBrowserProtocol3

[A3SB]: https://github.com/WoozyMasta/a2s/blob/master/pkg/a3sb/docs/README.md "Arma 3 Server Browser Protocol v3"
[CLI]: https://github.com/WoozyMasta/a2s/blob/master/CLI.md "Generated command-line reference"

[a2s-darwin-arm64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-darwin-arm64 "MacOS arm64 file"
[a2s-darwin-amd64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-darwin-amd64 "MacOS amd64 file"
[a2s-linux-amd64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-linux-amd64 "Linux amd64 file"
[a2s-linux-arm64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-linux-arm64 "Linux arm64 file"
[a2s-windows-amd64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-windows-amd64.exe "Windows amd64 file"
[a2s-windows-arm64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-windows-arm64.exe "Windows arm64 file"

[Support Me]: https://gist.github.com/WoozyMasta/7b0cabb538236b7307002c1fbc2d94ea
