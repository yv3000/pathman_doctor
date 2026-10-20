# Contributing to pathman

pathman is a Windows-only Go CLI. It reads and writes PATH through the
registry, so there is no server, database, container or hosted service in the
stack, and none should be added for a change to this repository.

## Prerequisites

- Windows 10 or Windows 11
- Go 1.26 or newer
- Git

## Setup

```powershell
git clone https://github.com/yv3000/pathman_doctor
cd pathman_doctor
go build ./...
go test ./...
```

Run the built binary with `go run . doctor` or `.\pathman.exe doctor` after
`go build -o pathman.exe .`.

## Checks before opening a pull request

| Task | Command |
|---|---|
| Build | `go build ./...` |
| Vet | `go vet ./...` |
| Tests | `go test ./... -cover` |
| Coverage gate | `./scripts/coverage-gate.ps1` |
| Lint | `golangci-lint run` |
| Vulnerabilities | `govulncheck ./...` |

CI runs the same commands on `windows-latest` for every push and pull request
to `main`. All three jobs (`build-test`, `lint`, `vulncheck`) must be green.

## Rules

- Every behaviour change ships with its tests in the same commit. A change
  without tests is not ready to merge.
- Keep commits focused: one fix or feature per commit, with the tests that prove
  it. Do not mix formatting, refactors and behaviour changes.
- Never write to the live `HKCU`/`HKLM` PATH from a test. Inject a fake
  `internal.Store` and let registry reads and writes go through the
  `openRegistryKey` seam. The only live check is `TestReadPATHUserIsReadable`,
  which is opt-in with `PATHMAN_LIVE_REGISTRY=1`.
- Keep `cmd` and `internal` at or above 80% statement coverage; the gate fails
  the build otherwise.
- Prefer the standard library. The CLI intentionally has two direct
  dependencies (`fatih/color` and `golang.org/x/sys`); justify any new one.
- Keep command output readable in a plain terminal and respect `NO_COLOR` via
  `fatih/color`.
- Back up both PATH scopes before a write and never write a scope that could not
  be read.

## Commit style

Use imperative subjects with a conventional prefix, for example:

```text
fix: surface System PATH read errors in fix command
feat(fix): back up PATH before writing changes
test(registry): cover registry key mapping for both PATH scopes
```

## Pull requests

1. Fork the repository and branch from `main`.
2. Make the change with tests and update `CHANGELOG.md` under `Unreleased`.
3. Confirm `go build ./...`, `go vet ./...`, `go test ./... -cover` and
   `golangci-lint run` pass locally.
4. Open the pull request and describe the behaviour change, not just the files.

Maintainers should enable branch protection for `main` and require the CI
checks before merging. This is a repository setting, not something code can
enforce.

## Releasing

1. Move the `Unreleased` entries in `CHANGELOG.md` to the new version.
2. Tag the commit as `vX.Y.Z` and push the tag.
3. The release workflow runs the tests, builds AMD64 and ARM64 binaries with the
   tag stamped into `main.version`, and attaches them to the GitHub release.
