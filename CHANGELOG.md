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

* `CLI`: `proxy` command
* `CLI`: Russian, German, Italian, Spanish, Czech,
  and Chinese localization with English fallback
* `CLI`: environment-variable configuration
* container image with the `a2s` CLI
* `a2s/server`: package for serving A2S queries over UDP
* `a2s/proxy`: package for cached and passthrough A2S proxying
* `a3sb`: binary rules encoding and A2S_RULES page generation
* `appid`: public curated Steam AppID registry for A2S-compatible games
* `a2s`: `QueryMeta` and `GetInfoWithMeta` for query transport metadata
* `a2s`: separate `AppID` and `GameID` fields with `Info.EffectiveID()`
* `a2s`: lossless logical `Packet` codec
* `a2s`: ordered, duplicate-preserving A2S_RULES entries
  with raw and parsed views

### Changed

* minimum supported Go version is now 1.25
* `CLI`: interactive tables adapt to terminal width
* `CLI`: mod and DLC tables use compact IDs when full URLs do not fit
* `CLI`: compact ping output and optional summary suppression
* `CLI`: strict command and argument validation
* `CLI`: `all --format json` emits one structured JSON document
* `CLI`: unified build and version metadata
* `a2s`: request and response types are separated
  into `QueryType` and `ResponseType`, replacing `Flag`
* `a2s`: query methods accept `context.Context`
* `a2s`: query latency covers complete logical requests,
  including challenge retries and split responses
* `a2s`: address-aware client constructors, options, accessors,
  and idempotent lifecycle management
* `a2s`: player query methods return slices directly
* `a2s` and `a3sb`: shared bounds-checked wire decoder
* `a2s`: default UDP receive buffer increased to 8192 bytes
* `a3sb`: automatic A2S/A3SB rules detection from a single A2S_RULES response

### Fixed

* `CLI`: generic A2S_INFO keywords are preserved in JSON output
* `CLI`: interrupted ping queries stop before final statistics are printed
* `a2s` and `a3sb`: malformed and truncated responses no longer cause panics
* `a2s`: reordered split responses are assembled correctly
* `a2s`: inconsistent split fragments are rejected
  and response allocations are bounded
* `a2s`: bounded challenge transactions
  and obsolete `GetChallenge` retry behavior
* `a2s`: concurrent queries on a single client are serialized safely
* `a2s`: SourceTV fields, server type and visibility keys,
  and signed player scores
* `a2s`: empty A2A_PING acknowledgements are accepted
* `a2s/proxy`: cache refresh and recovery behavior
* `a2s/proxy`: global and per-client rate-limit accounting
* `a2s/proxy`: bounded per-client rate-limit state
* `a2s/proxy`: fatal poller errors are reported
  instead of silently stopping refresh loops
* `a3sb`: deterministic DLC bit and hash ordering
* `a3sb`: DayZ-specific field parsing
* `a3sb`: rules page ordering and malformed page-set validation
* `keywords`: unknown enum and platform values are preserved

### Removed

* `CLI`: misleading `--format raw` mode
* `CLI`: redundant `--skip-info` rules option
* `a2s`: ambiguous `Info.ID` field
* `a2s`: transport-only `Info.Ping` field
* `github.com/woozymasta/steam` dependency in favor of the local AppID registry

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
