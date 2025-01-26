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

## [0.2.2][] - 2025-01-26

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
* appid package migrated to [woozymasta/steam](https://github.com/WoozyMasta/steam/tree/master/utils/appid)

[0.2.0]: https://github.com/WoozyMasta/a2s/compare/v0.1.0...v0.2.0

## [0.1.0][] - 2025-01-07

### Added

* First public release

[0.1.0]: https://github.com/WoozyMasta/a2s/tree/v0.1.0

<!--links-->
[Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html
