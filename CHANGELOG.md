# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.1]

### Added

- `-title` flag: raise the Alacritty window with the given title directly
  and exit, without the interactive picker; exits non-zero with an error on
  stderr when no window matches.
- `-info` flag: print the SQLite cache location and exit without launching
  the TUI.
- SQLite access-history cache that logs raised windows to a database in the
  user's cache directory.
- `access_history` table and migrations tooling (goose/v3, go-sqlite
  dependencies).
- ASCII art banner displayed at the top of the TUI (`banner.go`).

### Changed

- The raise and close AppleScript selectors now match the window name
  exactly (`name is`) instead of `name contains`, so a picked title that
  is a prefix of another window's title no longer raises or closes the
  wrong window.
- Windows in the UI are now sorted by most recently seen; windows without
  cached history keep the existing sort order.
- The picker now starts with the cursor on the window raised just before
  the last one, by swapping the first two recency-sorted entries in the
  display list.

### Fixed

- Cache scan now reads `accessed_at` as int64.
