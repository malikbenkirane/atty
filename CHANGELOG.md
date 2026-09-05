# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- SQLite access-history cache that logs raised windows to a database in the
  user's cache directory ([#1]).
- `access_history` table and migrations tooling (goose/v3, go-sqlite
  dependencies) ([#1]).

### Changed

- Windows in the UI are now sorted by most recently seen; windows without
  cached history keep the existing sort order ([#1]).

### Fixed

- Cache scan now reads `accessed_at` as int64 ([#1]).

[#1]: https://github.com/malikbenkirane/atty/issues/1
