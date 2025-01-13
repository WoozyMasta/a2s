<!-- omit in toc -->
# A2S

<!-- markdownlint-disable MD033 -->
<center>

![a2s-cli](winres/a2s-cli/icon64.png)
&emsp;
![a3sb-cli](winres/a3sb-cli/icon64.png)

`./a2s-cli` and `./a3sb-cli`
</center>
<!-- markdownlint-enable MD033 -->

Powerful command-line utilities and Go packages for querying Steam A2S
server information.
Built with specific support for Arma 3 and DayZ, these tools provide a
seamless way to retrieve essential server data.

* [Description](#description)
* [CLI Installation](#cli-installation)
* [A2S CLI](#a2s-cli)
* [A3SB CLI](#a3sb-cli)
* [Package](#package)
  * [A2S](#a2s-1)
  * [A3SB](#a3sb)
* [Protocol Documentation](#protocol-documentation)
* [Tested Games](#tested-games)
* [Support me 💖](#support-me-)

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

| Steam A2S Query `./a2s-cli` | A2S + Arma3/Dayz support  `./a3sb-cli` |
| --------------------------- | -------------------------------------- |
| [a2s-cli-darwin-arm64][]    | [a3sb-cli-darwin-arm64][]              |
| [a2s-cli-darwin-amd64][]    | [a3sb-cli-darwin-amd64][]              |
| [a2s-cli-linux-i386][]      | [a3sb-cli-linux-i386][]                |
| [a2s-cli-linux-amd64][]     | [a3sb-cli-linux-amd64][]               |
| [a2s-cli-linux-arm][]       | [a3sb-cli-linux-arm][]                 |
| [a2s-cli-linux-arm64][]     | [a3sb-cli-linux-arm64][]               |
| [a2s-cli-windows-i386][]    | [a3sb-cli-windows-i386][]              |
| [a2s-cli-windows-amd64][]   | [a3sb-cli-windows-amd64][]             |
| [a2s-cli-windows-arm64][]   | [a3sb-cli-windows-arm64][]             |

For Linux you can also use the command

```bash
# A2S
curl -#SfLo /usr/bin/a2s-cli \
  https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-cli-linux-amd64
chmod +x /usr/bin/a2s-cli
a2s-cli -h && a2s-cli -v

# A3SB
curl -#SfLo /usr/bin/a3sb-cli \
  https://github.com/WoozyMasta/a2s/releases/latest/download/a3sb-cli-linux-amd64
chmod +x /usr/bin/a3sb-cli
a3sb-cli -h && a3sb-cli -v
```

## A2S CLI

![logo-a2s][]

```txt
Description:
  CLI for querying Steam A2S server information.

Usage:
  a2s-cli [OPTIONS] <command> <host> <query port>

Example:
  a2s-cli ping 127.0.0.1 27016
  a2s-cli -j info 127.0.0.1 27016 | jq '.players'

Commands:
  info     Retrieve server information A2S_INFO;
  rules    Retrieve server rules A2S_RULES;
  players  Retrieve player list A2S_PLAYERS;
  all      Retrieve all available server information;
  ping     Ping the server with A2S_INFO.

Options:
  -j, --json               Output in JSON format;
  -r, --raw                Disable parse A2S_RULES values to types;
  -t, --deadline-timeout   Set connection timeout in seconds;
  -b, --buffer-size        Set connection buffer size;
  -c, --ping-count         Set the number of ping requests to send;
  -p, --ping-period        Set the period between pings in seconds;
  -t, --version            Show version, commit, and build time;
  -h, --help               Prints this help message.
```

<!-- markdownlint-disable MD033 -->
<details>
<summary>Examples of output for different commands (Click here to expand) 👈</summary>
<!-- markdownlint-enable MD033 -->

<!-- omit in toc -->
### Info

```bash
./a2s-cli info 127.0.0.1:27016
```

```txt
Property         Value
==============================================================================
Query type:      Source
Protocol:        17
Server name:     Some CSS server
Map on server:   de_dust2
Game folder:     cstrike
Game name:       Counter-Strike: Source
Steam AppID:     240
Players/Slots:   64/65
Bots count:      0
Server type:     Dedicated
Server OS:       Linux
Need password:   false
VAC protected:   true
Game version:    6630498
Port:            27016
Server SteamID:  85568392924000000
Keywords:        increased_maxplayers, startmoney, alltalk, 100, 66 de_dust2
                 dust2
Server ping:     30 ms
==============================================================================
A2S_INFO response for 127.0.0.1:27016
```

<!-- omit in toc -->
### Rules

```bash
./a2s-cli rules 127.0.0.1:27016
```

```txt
Rule                                 Value
============================================
bot_quota                            0
command_buffer_version               2.8
coop                                 0
custom_chat_colors_version           3.0.1
...
tv_enable                            1
=============================================
A2S_RULES response for 127.0.0.1:27016

```

<!-- omit in toc -->
### Players

```bash
./a2s-cli players 127.0.0.1:27016
```

```txt
  #  PlayTime  Score  Name
==================================
  1  6h36m10s  51     John Doe
  2  6h19m37s  0      Jane Smith
  3  3h44m30s  26     Alex Johnson
...
 62  2m25s     3      Chris Taylor
 63  1m19s     0      Emily Davis
 64  25s       0      Pat Brown
===================================
A2S_PLAYERS response for 127.0.0.1:27016
```

<!-- omit in toc -->
### Ping

```bash
./a2s-cli ping -c 5 127.0.0.1:27016
```

```txt
Start 5 times ping 127.0.0.1:27016 with 1s period

A2S_INFO response server=127.0.0.1:27016 folder="cstrike" name="Some CSS server" time=29.5289ms
A2S_INFO response server=127.0.0.1:27016 folder="cstrike" name="Some CSS server" time=29.274ms
A2S_INFO response server=127.0.0.1:27016 folder="cstrike" name="Some CSS server" time=29.0262ms
A2S_INFO response server=127.0.0.1:27016 folder="cstrike" name="Some CSS server" time=29.6299ms
A2S_INFO response server=127.0.0.1:27016 folder="cstrike" name="Some CSS server" time=28.7014ms

Transmitted 5 request, received 5 response, failed 0
Min=28.7014ms Max=29.6299ms Avg=29.23208ms
```

</details>

## A3SB CLI

![logo-a3sb][]

```txt
Description:
  CLI for querying A2S server information and working with A3SB subprotocol for Arma 3 and DayZ.

Usage:
  a3sb-cli [OPTIONS] <command> <host> <query port>

Example:
  a3sb-cli ping 127.0.0.1 27016
  a3sb-cli -j info 127.0.0.1 27016 | jq '.players'

Commands:
  info     Retrieve server information A2S_INFO;
  rules    Retrieve server rules A2S_RULES;
  players  Retrieve player list A2S_PLAYERS;
  all      Retrieve all available server information;
  ping     Ping the server with A2S_INFO.

Options:
  -j, --json               Output in JSON format;
  -i, --app-id             AppID for more accurate results;
  -t, --deadline-timeout   Set connection timeout in seconds;
  -b, --buffer-size        Set connection buffer size;
  -c, --ping-count         Set the number of ping requests to send;
  -p, --ping-period        Set the period between pings in seconds;
  -t, --version            Show version, commit, and build time;
  -h, --help               Prints this help message.
```

<!-- markdownlint-disable MD033 -->
<details>
<summary>Examples of output for different commands (Click here to expand) 👈</summary>
<!-- markdownlint-enable MD033 -->

<!-- omit in toc -->
### Info Arma 3

```bash
./a2s-cli -i 107410 info 127.0.0.1:27016
```

```txt
Property         Value
===============================================
Query type:      Source
Protocol:        17
Server name:     Some Arma 3 Server
Map on server:   Altis
Game folder:     Arma3
Game name:       Some Arma 3 Server Description
Steam AppID:     107410
Players/Slots:   112/115
Bots count:      0
Server type:     Dedicated
Server OS:       Windows
Need password:   false
VAC protected:   false
Game version:    2.18.152405
Port:            2302
Server SteamID:  90241503198000000
Server ping:     175 ms
===============================================
A2S_INFO response for 127.0.0.1:27016
```

<!-- omit in toc -->
### Info DayZ

```bash
./a2s-cli -i 221100 info 127.0.0.1:27016
```

```txt
Property             Value
=================================================
Query type:          Source
Protocol:            17
Server name:         Some DayZ Server
Map on server:       sakhal
Game folder:         dayz
Game name:           Some DayZ Server Description
Steam AppID:         221100
Players/Slots:       29/80
Bots count:          0
Server type:         Dedicated
Server OS:           Linux
Need password:       false
VAC protected:       true
Game version:        1.26.159040
Port:                2402
Server SteamID:      90241641813000000
Shard:               123ABC
In game time:        18h6m0s
Time day x:          8.000000
Time night x:        2.000000
Game port:           0
Players queue:       0
BattlEye protected:  true
Third person:        true
External:            true
Private hive:        true
Modded:              true
Whitelist:           false
Fle patching:        false
Need DLC:            false
Server ping:         22 ms
=================================================
A2S_INFO response for 127.0.0.1:27016
```

<!-- omit in toc -->
### Rules Arma 3

```bash
./a2s-cli -i 107410 info 127.0.0.1:27016
```

```txt

  #  DLC Name              DLC URL
=======================================================================
  1  Contact (Platform)    https://store.steampowered.com/app/1021790
  2  Karts                 https://store.steampowered.com/app/288520
...
 11  Jets                  https://store.steampowered.com/app/601670
 12  Laws of War           https://store.steampowered.com/app/571710
=======================================================================

  #  Creator DLC Name             Creator DLC URL
==============================================================================
  1  Creator DLC: Western Sahara  https://store.steampowered.com/app/1681170
==============================================================================

  #  Mod Name                                     Mod URL
=====================================================================================================================
  1  Zeus Enhanced 1.15.1                         https://steamcommunity.com/sharedfiles/filedetails/?id=1779063631
  2  Zeus Additions                               https://steamcommunity.com/sharedfiles/filedetails/?id=2387297579
...
 47  3CB Factions                                 https://steamcommunity.com/sharedfiles/filedetails/?id=1673456286
=====================================================================================================================
A2S_RULES response for 144.76.173.27:2393
```

<!-- omit in toc -->
### Rules DayZ

```bash
./a2s-cli -i 221100 info 127.0.0.1:27016
```

```txt
Option             Value
===============================================
Description:       Some DayZ Server Description
Allowed build:     0
Client port:       0
Dedicated:         false
Island:            sakhal
Language:          Russian
Platform:          Linux
Required build:    0
Required version:  126
TimeLeft:          15
================================================

  #  DLC Name    DLC URL
=============================================================
  1  Frost Line  https://store.steampowered.com/app/2968040
=============================================================

  #  Mod Name             Mod URL
=============================================================================================
  1  VPPAdminTools        https://steamcommunity.com/sharedfiles/filedetails/?id=1828439124
  2  Community Framework  https://steamcommunity.com/sharedfiles/filedetails/?id=1559212036
=============================================================================================
A2S_RULES response for 127.0.0.1:27016
```

<!-- omit in toc -->
### Players Arma 3

```bash
./a2s-cli -i 107410 info 127.0.0.1:27016
```

```txt
  #  PlayTime            Score       Name
=================================================
  1  8h47m34.482421875s  0           John Doe
  2  8h7m26.66796875s    0           Jane Smith
...
111  30.060682297s       0           Chris Taylor
112  12.609279633s       0           Emily Davis
=================================================
A2S_PLAYERS response for 127.0.0.1:27016
```

</details>

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

During development, the functionality of `a2s-cli` and `a3sb-cli` was
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

## Support me 💖

If you enjoy my projects and want to support further development,
feel free to donate! Every contribution helps to keep the work going.
Thank you!

<!-- omit in toc -->
### Crypto Donations

<!-- cSpell:disable -->
* **BTC**: `1Jb6vZAMVLQ9wwkyZfx2XgL5cjPfJ8UU3c`
* **USDT (TRC20)**: `TN99xawQTZKraRyvPAwMT4UfoS57hdH8Kz`
* **TON**: `UQBB5D7cL5EW3rHM_44rur9RDMz_fvg222R4dFiCAzBO_ptH`
<!-- cSpell:enable -->

Your support is greatly appreciated!

<!-- Links -->
[logo-a2s]: assets/a2s.png
[logo-a3sb]: assets/a3sb.png

[Steam Server Queries]: https://developer.valvesoftware.com/wiki/Server_queries
[Arma 3 Server Browser Protocol v3]: https://community.bistudio.com/wiki/Arma_3:_ServerBrowserProtocol3
[A3SB Protocol v3 Specification 🇬🇧]: https://github.com/WoozyMasta/a2s/blob/master/pkg/a3sb/docs/README.md "🇬🇧"
[A3SB Protocol v3 Specification 🇷🇺]: https://github.com/WoozyMasta/a2s/blob/master/pkg/a3sb/docs/README_ru.md "🇷🇺"

[a2s-cli-darwin-arm64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-cli-darwin-arm64 "MacOS arm64 file"
[a2s-cli-darwin-amd64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-cli-darwin-amd64 "MacOS amd64 file"
[a2s-cli-linux-i386]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-cli-linux-386 "Linux i386 file"
[a2s-cli-linux-amd64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-cli-linux-amd64 "Linux amd64 file"
[a2s-cli-linux-arm]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-cli-linux-arm "Linux arm file"
[a2s-cli-linux-arm64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-cli-linux-arm64 "Linux arm64 file"
[a2s-cli-windows-i386]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-cli-windows-386.exe "Windows i386 file"
[a2s-cli-windows-amd64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-cli-windows-amd64.exe "Windows amd64 file"
[a2s-cli-windows-arm64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a2s-cli-windows-arm64.exe "Windows arm64 file"

[a3sb-cli-darwin-arm64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a3sb-cli-darwin-arm64 "MacOS arm64 file"
[a3sb-cli-darwin-amd64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a3sb-cli-darwin-amd64 "MacOS amd64 file"
[a3sb-cli-linux-i386]: https://github.com/WoozyMasta/a2s/releases/latest/download/a3sb-cli-linux-386 "Linux i386 file"
[a3sb-cli-linux-amd64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a3sb-cli-linux-amd64 "Linux amd64 file"
[a3sb-cli-linux-arm]: https://github.com/WoozyMasta/a2s/releases/latest/download/a3sb-cli-linux-arm "Linux arm file"
[a3sb-cli-linux-arm64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a3sb-cli-linux-arm64 "Linux arm64 file"
[a3sb-cli-windows-i386]: https://github.com/WoozyMasta/a2s/releases/latest/download/a3sb-cli-windows-386.exe "Windows i386 file"
[a3sb-cli-windows-amd64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a3sb-cli-windows-amd64.exe "Windows amd64 file"
[a3sb-cli-windows-arm64]: https://github.com/WoozyMasta/a2s/releases/latest/download/a3sb-cli-windows-arm64.exe "Windows arm64 file"
