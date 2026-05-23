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

### Changed

* `a2s` query methods accept `context.Context` for cancellation
  and total query deadlines
* `a2s` report complete logical query latency consistently
  across challenge retries and split responses
* `a2s` player query methods return slices directly
  instead of pointers to slices
* `a2s` expose raw and parsed A2S_RULES views
* `a2s` replace legacy client construction with address-aware constructors,
  options, accessors, and idempotent lifecycle management

### Fixed

* `a3sb` assemble one-based rule pages by page number
  and reject inconsistent, missing, or conflicting pages
* `a2s` and `a3sb` reject malformed and truncated UDP responses without panics
* `a2s` assemble reordered split responses using metadata from fragment zero
* `a2s` reject inconsistent split fragments and bound response allocations
* `a2s` handle challenge responses through bounded transactions
  and keep deprecated `GetChallenge` path from retrying its final response
* `a2s` serialize concurrent query transactions on one client
* `a3sb` preserve deterministic DLC bit and hash ordering
* `a3sb` parse the DayZ `dedicated` rule according to its wire value
* `a2s` align response models with A2S wire and JSON contracts,
  including SourceTV fields, server type/visibility keys, and signed scores
* `keywords` preserve unknown enum values
  instead of mapping them to known defaults

### Removed

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
