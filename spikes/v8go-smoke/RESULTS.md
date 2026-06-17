# v8go Spike Result Guide

This file defines how to interpret the first real CI results for the Fase 1A
viability spike.

## Expected Artifacts

Supported platforms:

- `v8go-smoke-ubuntu-latest-artifacts`
- `v8go-smoke-macos-latest-artifacts`

Windows evidence:

- `v8go-smoke-windows-artifacts`

Expected files:

- `cold-start-ubuntu-latest.txt`
- `cold-start-macos-latest.txt`
- `cold-start-windows-latest.txt`
- `build.log`

## Success Matrix

### Linux / macOS

Treat the platform as provisionally viable only if all of the following are true:

1. `go run .` succeeds and prints `eval=42 promise=42`.
2. `go test ./eventloop -v` passes.
3. `go test ./bench -run TestDoesNotExist -bench . -benchmem -cpu 1 -count 10`
   completes and produces both `BenchmarkColdStartGoja` and
   `BenchmarkColdStartV8Go`.

Interpretation:

- If all three pass, the platform is good enough for deeper runtime research.
- If smoke passes but event loop fails, V8 embedding is not yet operational for
  async/native integration.
- If benchmarks do not include `BenchmarkColdStartV8Go`, the platform is not yet
  ready for fair engine comparison.

### Windows

Treat Windows as still blocked unless all of the following become true:

1. `go build -x .` exits successfully.
2. The spike binary becomes runnable.
3. `BenchmarkColdStartV8Go` can execute locally or in CI.

Current expected interpretation:

- `build.log` still shows the external linker failure.
- `cold-start-windows-latest.txt` contains only the `goja` baseline.
- This preserves the short-term no-go in ADR-001.

## Cold Start Comparison Rules

Use the archived benchmark outputs, not a single copied console line.

Review:

1. Range and stability across the 10 samples.
2. `ns/op` for `goja` vs `v8go`.
3. `B/op` and `allocs/op`.

Decision guidance:

- If `v8go` is materially slower at cold start and does not unlock a stable
  cross-platform path, keep `goja`.
- If `v8go` is slower at cold start but wins on later throughput-oriented phases,
  keep it under investigation only.
- If `v8go` is comparable enough and the event-loop spike is stable on Linux and
  macOS, proceed to deeper API-integration experiments.

## Exit Criteria For Fase 1A

Mark the pending checks as complete only when:

1. Linux x64 and macOS results exist and are archived.
2. The event-loop spike passes on those supported platforms.
3. A first cross-platform cold-start comparison is written down in ADR-001 or a
   follow-up ADR/update.
4. Windows is either unblocked or explicitly recorded as a continuing blocker.
