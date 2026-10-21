# Changelog

All notable changes to pathman are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

Nothing yet.

## [1.3.0] - 2026-10-21

### Added

- `--entry` values are validated at the flag boundary: empty, whitespace-only
  and semicolon-separated values exit with status 2 before PATH is read.
  `internal.ValidateEntry` is the single rule every command uses.
- CI enforces coverage. `scripts/coverage-gate.ps1` fails the build when `cmd`
  or `internal` drops below 80% or the module total drops below 80%.
- The registry layer is injectable, so reads, writes, missing values and error
  propagation are tested against a fake key instead of the live hive.
- `internal.ValidateOnlyFlag` owns the `--only` rule next to `ValidateEntry`, so
  both flags are validated in one place instead of inside the parser.
- `PATHMAN_LOG=1` writes leveled `INFO`/`WARN`/`ERROR` diagnostics from
  `internal.Logf`. The default run stays silent, so the console output is
  unchanged.
- CI uploads `coverage.out` as the `coverage` artifact on every run, including
  failed ones, so a regression can be read per function.
- CI runs the full suite weekly with the Go module cache disabled, which proves
  a cold clone builds and tests.

### Changed

- The live registry smoke test is opt-in (`PATHMAN_LIVE_REGISTRY=1`); the
  default suite runs without touching the registry.
- Dependabot groups routine Go module and GitHub Actions updates into one pull
  request per ecosystem per week.
- CI pins `govulncheck` to v1.9.0 instead of `@latest`.

## [1.2.2] - 2026-10-18

### Changed

- The deferred uninstall cleanup now tries to break away from a hosting job
  object before falling back to a plain detached process, so the folder is
  removed in more environments (agent sandboxes, CI runners).

## [1.2.1] - 2026-10-18

### Fixed

- `uninstall` now schedules the deletion of its own folder for when the process
  exits. Windows locks the running executable, so the first live run removed the
  PATH entry but could not delete the files.

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
