# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

`coub-dl` is a small Go CLI that downloads videos from coub.com. Two subcommands:
`download <link|permalink>` (one coub, public) and `sync` (all liked coubs via
`API_TOKEN`). It shells out to `ffmpeg` to loop the video under the audio.

## Commands

```sh
go build -o coub-dl .          # build
go test ./...                  # all tests (includes a ~3s retry-backoff test)
go test -short ./...           # skip the slow retry test
go test -run TestBestURL ./internal/coub/   # single test
go vet ./...                   # vet
gofmt -w .                     # format

go run . download 2uywin                 # run download
API_TOKEN=xxx go run . sync -workers 10  # run sync
```

stdlib `flag` requires flags **before** the positional link
(`download -out dir <link>`, not `download <link> -out dir`).

## Architecture

Two layers:

- **`package main`** (root: `main.go`, `commands.go`) — CLI: arg parsing,
  subcommand dispatch, exit codes (`0`/`1`/`64`/`130`), and the `sync` worker pool.
- **`internal/coub`** — the domain: API client plus the download/mux pipeline.

Key flows that span files:

- **`Client`** (`internal/coub/client.go`) holds the `*http.Client` + token and
  exposes `Get`, `FetchLikes`, `LikesCount`, `Download`. Every HTTP call funnels
  through `doGet` in `retry.go`, which retries transient failures (network, 429,
  5xx) with ctx-aware exponential backoff — jittered, and honoring a `Retry-After`
  header when present. `ctx` is always a per-call parameter, never a struct field.
- **`FetchLikes`** is a producer goroutine streaming coubs over a buffered
  channel, with a second channel for a single error.
- **`sync`** (`commands.go`): `syncer.run` consumes that channel through a
  bounded worker pool (WaitGroup + semaphore), dedups permalinks (pagination can
  repeat), and hands results to one `report` goroutine that owns the counters
  and stderr progress. Workers do not print or touch shared counters.
- **`Download`** (`download.go`): picks the best available quality, downloads
  video+audio to temp files, then a single ffmpeg pass
  (`-stream_loop -1 … -shortest`, stream copy) loops and muxes them. The ffmpeg
  shell-out is isolated in `mux.go`.

## Conventions and decisions

- **stdlib-only** — no third-party deps (rejected go-flags, lgr). No interfaces;
  every type has one implementation.
- **Output name = `<permalink>.mp4`** — used as-is, not sanitized; but unsafe
  permalinks or `-name` values (`..`, path separators, a leading `-`) are
  rejected, not cleaned. Title/tags/author go into the mp4 metadata, not the
  filename. `sync` lays files out as `outDir/YYYY/MM/`; `download` is flat.
- **Single ffmpeg pass with `-c copy`** (no transcode) for portability. Large
  output is the looped video baked to the audio length — by design, not a bug.
- Errors are values (no sentinels). `ffmpeg` is required at runtime and checked
  up front in each command.
- Tests sit beside sources (`*_test.go`, white-box `package coub`); HTTP paths
  are tested with `httptest` by passing the test URL straight to `doGet`.

## Docs

- `README.md` — install and usage.
- `docs/COUB_API.md` — coub.com API reference (endpoints, data shapes).
