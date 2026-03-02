# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

Build uses [Task](https://taskfile.dev/) (`Taskfile.yml`):

```sh
task build        # fmt + production build (CGO_ENABLED=0, trimpath, PGO, static binary)
task build-debug  # fmt + debug build with race detector (CGO_ENABLED=1)
task lint         # fmt + golangci-lint (5m timeout)
task fmt          # gci write + gofumpt + betteralign
task update       # go get -u + go mod tidy
task release      # goreleaser release --clean -p 4
```

Run tests:

```sh
go test ./...                   # all packages
go test .                       # main package only (calendar_test.go)
go test ./geoip/...             # geoip package
go test ./ics/...               # ics package
go test ./oauth/...             # oauth package
go test -run TestParseCalendarEvent ./...  # run a specific test
```

Tests use `net/http/httptest` to stub HTTP endpoints; no live network required. `oauth/oauth_test.go` is in `package oauth` (white-box) to access unexported `tokenFromFile` and `saveToken`. All other test packages use the `_test` suffix for black-box testing.

Linting uses golangci-lint v2 with `default: all` and a set of disabled linters (see `.golangci.yml`). Formatters enforced: `gci`, `gofmt`, `gofumpt`, `goimports`.

The build embeds `ldflags` version metadata: `GitTag`, `GitCommit`, `GitDirty`, `BuildTime` into `main.*` variables.

## Architecture

Single-binary Go CLI (`package main`) with three internal subpackages.

**Entry point — `main.go`**
Parses CLI flags via `peterbourgon/ff/v4` (supports `--config` YAML file and `IMB_` env var prefix). Loads embedded Google OAuth2 client credentials from `assets/credentials.json` (embedded at build time via `//go:embed`). Initializes the Google Calendar API client, then runs two goroutines concurrently behind a channel + `time.Timer` API timeout:
- One goroutine fetches public holidays (`getHolidayEvents`)
- One goroutine fetches calendar events, waits for holidays, then prints stats

**Core logic — `calendar.go`**
- `getCalendarID`: resolves a human-readable calendar name to a Google Calendar ID (paginated)
- `getCalendarEvents`: fetches events in the date range, optionally filters by prefix (`--search`), skips recurring events unless `--recurring`, accumulates hours per day into `map[string]workEvent`
- `printMonthlyStats`: prints sorted daily work log with totals, then cross-references against holiday dates to warn of holiday work
- `getHolidayEvents`: orchestrates the GeoIP → ICS pipeline (non-fatal; returns empty map on any error)

**Subpackages:**

| Package | File | Purpose |
|---------|------|---------|
| `oauth` | `oauth/oauth.go` | OAuth2 token lifecycle: load from `token.json`, auto-refresh if expired, or launch interactive browser flow with a local HTTP callback server (Chi router, random free port, UUID v7 CSRF state) |
| `geoip` | `geoip/ifconfig.go` | HTTP client for `ifconfig.co/json` — returns country ISO code for the current public IP |
| `ics` | `ics/ics.go` | HTTP client for `officeholidays.com` ICS feed by country code; implements `goics.Consumer` to decode calendar events |

**Runtime files:**
- `assets/credentials.json` — Google OAuth2 **app** credentials (embedded in binary; do not remove from `assets/`)
- `token.json` — per-user OAuth2 token (auto-created/refreshed at runtime; written atomically via `google/renameio`)

**Key design constraints:**
- All errors in the main flow are fatal (`log.Fatalf`) — this is a CLI tool, not a library
- `CGO_ENABLED=0` for production builds (fully static binary)
- Holiday lookup is best-effort: any failure silently returns an empty map
- Event duration is rounded to whole hours (`workDuration.Round(time.Hour).Hours()`)
