# Changelog

All notable changes to pathman are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.2.0] - 2026-10-18

### Added

- Test suite covering PATH parsing, analysis, removal planning, registry key
  mapping, backups and every command.
- Windows CI that builds, runs `go vet`, tests with a coverage report, runs
  golangci-lint and scans dependencies with govulncheck.
- Weekly Dependabot updates for Go modules and GitHub Actions.
- `CONTRIBUTING.md`, `CHANGELOG.md` and `LICENSE`.
- A backup of the untouched PATH value before every write, stored under
  `%LOCALAPPDATA%\pathman\backups`.

### Changed

- `pathman fix` reports partial write failures and exits with status 1 when a
  scope could not be updated.
- `pathman fix` validates its flags; unknown flags, stray arguments and invalid
  `--only` values exit with status 2 without touching PATH.
- `pathman uninstall` removes its own User PATH entry and files instead of
  executing a remote PowerShell script.
- Release builds inject the tag version, so `pathman --version` matches the
  published release.
- Dependencies updated to `fatih/color` 1.19.0 and `golang.org/x/sys` 0.49.0;
  the minimum Go version is now 1.26.
- GitHub Actions updated to `actions/checkout@v7`, `actions/setup-go@v7` and
  `softprops/action-gh-release@v3`.

### Fixed

- `%VAR%` entries such as `%SystemRoot%\system32` were treated as missing
  folders, which could make `fix` delete working System PATH entries.
- System PATH read errors in `fix` were silently ignored.
- `go vet` flagged an unescaped percent sign in the uninstall hint.
- CI failed on Windows because Go sources were checked out with CRLF line
  endings; `.gitattributes` now pins them to LF and tests print per-package
  coverage without a profile file.

### Removed

- The committed `pathman-windows-amd64.exe` build artifact.

## [1.1.0] - 2026-06-22

### Added

- `--only dead`, `--only dup` and `--entry` flags for `fix`.
- PowerShell installer that detects the architecture and updates the User PATH.
- PATH change broadcasting so open terminals pick up fixes without a restart.

## [1.0.0] - 2026-06-11

### Added

- `doctor` command that reports dead and duplicate PATH entries.
- `fix` command with a confirmation prompt before writing PATH.
- PowerShell uninstaller and registry-backed System/User PATH handling.
