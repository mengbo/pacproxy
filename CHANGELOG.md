# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-02-03

### Added

- Initial release of pacproxy
- PAC (Proxy Auto-Config) file support with automatic fetching and caching
- HTTP proxy server with CONNECT method support for HTTPS tunneling
- Multiple upstream proxy support: DIRECT, HTTP, SOCKS5
- JavaScript PAC engine using goja for executing FindProxyForURL
- CLI with cobra framework supporting custom listen address and PAC URL
- Configurable logging levels (debug, info, warn, error)
- Graceful shutdown on interrupt signals
- Concurrent connection handling
- Default configuration: listens on 127.0.0.1:1080, uses clash-server PAC
- Usage help displayed on startup showing environment variable configuration

### Features

- Fetches PAC files from HTTP/HTTPS URLs with caching
- Parses PAC proxy directives: DIRECT, PROXY, SOCKS5, SOCKS4
- Supports PAC fallback chains (e.g., "PROXY p1:8080; PROXY p2:8080; DIRECT")
- Implements standard PAC helper functions:
  - `isPlainHostName()`
  - `dnsDomainIs()`
  - `localHostOrDomainIs()`
  - `isResolvable()`
  - `isInNet()`
  - `dnsResolve()`
  - `myIpAddress()`
  - `dnsDomainLevels()`
  - `shExpMatch()`
- HTTP CONNECT method handling for HTTPS proxying
- Bidirectional data tunneling for CONNECT requests
- Thread-safe PAC engine with mutex protection

[Unreleased]: https://github.com/mengbo/pacproxy/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/mengbo/pacproxy/releases/tag/v0.1.0
