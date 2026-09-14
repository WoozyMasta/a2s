<!-- markdownlint-disable MD024 -->
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][],
and this project adheres to [Semantic Versioning][].

<!--
## Unreleased

### Added
### Changed
### Removed
-->

## Unreleased

### Added

* `CLI` adds the `proxy` command
* `CLI` localizes in Russian, German, Italian, Spanish, Czech, and Chinese
  with English fallback
* `a2s/server` provides a UDP A2S server
  with secure challenge-by-default handling,
  bounded Source/GoldSource packetization, controlled shutdown,
  panic isolation, atomic snapshots, middleware, and proxy composition
* `a2s/proxy` provides reusable cached response storage, polling,
  optional global and per-client rate limits, bounded live passthrough,
  local challenge handling, and recovery lifecycle
* `appid` exposes the curated Steam AppID registry for A2S-compatible games
* `a2s` adds `QueryMeta` and `GetInfoWithMeta` for query transport metadata
* `a2s` exposes protocol-faithful `AppID` and `GameID` fields,
  with `Info.EffectiveID()` for effective game ID lookup
* `a2s` exposes a lossless logical `Packet` codec for single-packet responses

### Changed

* project requires Go 1.25 or newer
* `CLI` adapts interactive tables to terminal width
* `CLI` uses strict command/argument validation and unified version metadata
* `a2s` separate request and response type bytes
  into `QueryType` and `ResponseType`, removing the ambiguous `Flag` API
* `a2s` expose ordered, duplicate-safe A2S_RULES entries through `Rules`,
  with explicit lossy map conversion; `a3sb` preserves the same semantics
* `a2s` and `a3sb` replace the legacy binary reader with a shared,
  bounds-checked wire decoder across protocol parsers,
  preserving malformed packet handling and improving parser performance
* `a2s` uses an 8192-byte default UDP receive buffer
  shared by A2S and A3SB queries without mutating client configuration
* `a2s` query methods accept `context.Context` for cancellation
  and total query deadlines
* `a2s` report complete logical query latency consistently
  across challenge retries and split responses
* `a2s` player query methods return slices directly
  instead of pointers to slices
* `a2s` expose raw and parsed A2S_RULES views
* `a2s` replace legacy client construction with address-aware constructors,
  options, accessors, and idempotent lifecycle management
* `a3sb` automatically detects native A2S rules and known A3SB versions
  from one rules response without an additional A2S_INFO request

### Fixed

* `CLI` preserve generic A2S_INFO keywords in JSON output
* `CLI` stops an active ping query on interrupt before printing final statistics
* `a2s` and `a3sb` reject malformed and truncated UDP responses without panics
* `a2s` assemble reordered split responses using metadata from fragment zero
* `a2s` reject inconsistent split fragments and bound response allocations
* `a2s` handle challenge responses through bounded transactions
  and keep obsolete `GetChallenge` path from retrying its final response
* `a2s` serialize concurrent query transactions on one client
* `a2s` align response models with A2S wire and JSON contracts,
  including SourceTV fields, server type/visibility keys, and signed scores
* `a2s` accepts empty A2A_PING acknowledgements
  while retaining textual payload support
* `a3sb` preserve deterministic DLC bit and hash ordering
* `a3sb` parse the DayZ `dedicated` rule according to its wire value
* `a3sb` assemble one-based rule pages by page number
  and reject inconsistent, missing, or conflicting pages
* `keywords` preserve unknown enum and platform values
  instead of mapping them to known defaults

### Removed

* `CLI` removes the misleading `--format raw` output mode
* `a2s` removes the ambiguous `Info.ID` field
* `a2s` removes transport-only `Info.Ping` from the A2S_INFO model and JSON
* remove the redundant `--skip-info` rules option
  after rules detection stopped requiring `A2S_INFO`
* remove the `github.com/woozymasta/steam` dependency
  in favor of a curated local A2S AppID registry

## [0.3.2][] - 2026-03-07

### Added

* `a2s` exported `ErrChallengeLoop` and `ErrQueryUnsupported` for clearer
  handling of non-standard `A2S_PLAYER` / `A2S_RULES` responses

### Fixed

* `a2s` tolerate truncated GoldSource `A2S_INFO` responses where trailing
  `VAC` and/or `bots` bytes are missing
* `a2s` improve UDP response handling by skipping unrelated or truncated
  datagrams during packet reads and split-packet assembly

### Changed

* `a2s` return clearer diagnostics for unsupported queries
 (e.g. `A2S_INFO` returned for `rules`) and challenge-loop behavior

[0.3.2]: https://github.com/WoozyMasta/a2s/compare/v0.3.1...v0.3.2

## [0.3.1][] - 2026-01-31

### Added

* `a2s` support bz2-compressed multi-packet responses
* `a2s` support split (multi-packet) responses
  (rewritten for correctness and reliability)

### Fixed

* `a2s` avoid panics on truncated UDP reads and
  tolerate missing EDF byte in `A2S_INFO`

### Changed

* `a2s` increased default read buffer size to 4096 bytes
  to reduce UDP truncation risk

[0.3.1]: https://github.com/WoozyMasta/a2s/compare/v0.3.0...v0.3.1

## [0.3.0][] - 2025-12-30

### Added

* `a2s` package now exports 50+ specific error types for detailed error handling
* `a2s` CLI unified `a2s-cli` and `a3sb-cli` into single `a2s`
  command with subcommands

### Changed

* `a2s` and `a3sb` reduced memory allocations and memory usage by up
  to 3-4x through buffer reuse
* `a2s` improved error messages with context using `errors.Join`

[0.3.0]: https://github.com/WoozyMasta/a2s/compare/v0.2.3...v0.3.0

## [0.2.3][] - 2025-08-26

### Added

* `a3sb` DayZ Badlands DLC ID and AppID

### Changed

* updated golang ci lint and fix exported code comments
* updated direct dependencies

[0.2.3]: https://github.com/WoozyMasta/a2s/compare/v0.2.2...v0.2.3

## [0.2.2][] - 2025-01-26

### Added

* `a3sb-cli` show keywords for Arma3 in info request

### Changed

* `keywords` fix parse for empty string

[0.2.2]: https://github.com/WoozyMasta/a2s/compare/v0.2.1...v0.2.2

## [0.2.1][] - 2025-01-13

### Added

* `a2s` add `NewWithString()` new for init connection by `ip:port` string
* `a2s-cli` and `a3sb-cli` ability to specify both the `host` and `port` as
  separate arguments or as one `host:port`

### Changed

* `a2s-cli` fixed json output

[0.2.1]: https://github.com/WoozyMasta/a2s/compare/v0.2.0...v0.2.1

## [0.2.0][] - 2025-01-13

Refactoring and Simplification

### Added

* `keywords` new parser for Arma3
* `keywords` tests for parsing
* `keywords/types` package with new types for Arma3 `GameType`,
  `ServerLang` and `Platform`
* missed documentation for packages
* `a3sb-cli` ping support
* `a2s` new function `NewWithAddr()` use `*net.UDPAddr` as argument

### Changed

* moved `ServerLang` struct to package `pkg/keywords/types`
* moved packages `bread`, `tableprinter` to internal
* ping ring buffer separate as internal package
* cli args parse now with `jessevdk/go-flags`
* `a2s` function `CreateClient()` replaced with `Create()` and use
  `*net.UDPAddr` as argument

### Removed

* heavy and unnecessary logging packages and CLI parameter parsing
* appid package migrated to
  [woozymasta/steam](https://github.com/WoozyMasta/steam/tree/master/utils/appid)

[0.2.0]: https://github.com/WoozyMasta/a2s/compare/v0.1.0...v0.2.0

## [0.1.0][] - 2025-01-07

### Added

* First public release

[0.1.0]: https://github.com/WoozyMasta/a2s/tree/v0.1.0

<!--links-->
[Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html
