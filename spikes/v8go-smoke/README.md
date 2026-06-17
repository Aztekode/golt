# v8go Smoke Test

This spike isolates a minimal V8 binding experiment from the main Golt module.

## Goal

Validate the minimum capabilities needed before attempting a real engine
migration:

- Create an isolate and context
- Evaluate a simple script
- Set a global value from Go
- Execute a Promise and flush microtasks

## Run

From this directory:

```bash
go mod tidy
go run .
```

Expected output on a working platform:

```text
eval=42 promise=42
```

## Benchmark

Cold-start benchmarking is isolated under `./bench` so the `goja` baseline can run
even when the main `v8go` smoke binary does not link on Windows.

Supported commands:

```bash
go test ./bench -run TestDoesNotExist -bench . -benchmem -cpu 1 -count 10
```

Windows baseline only:

```bash
go test ./bench -run TestDoesNotExist -bench BenchmarkColdStartGoja -benchmem -cpu 1 -count 10
```

The benchmark intentionally measures a fresh VM/context creation plus a first
script evaluation, which is enough for early engine-selection evidence.

Measurement rules for this spike:

- use `-cpu 1` to reduce scheduler noise during early engine comparison
- use `-count 10` so the output can be compared across runs instead of trusting
  a single sample
- archive raw benchmark output per platform before drawing conclusions

## Event Loop Spike

The hybrid event-loop experiment lives under `./eventloop` and models the design
sketched in `PLAN.md`: Go goroutines enqueue tasks, a `time.Ticker` pumps V8
microtasks, and all V8 access stays serialized through the same context owner.

Supported command on platforms where `v8go` links:

```bash
go test ./eventloop -v
```

The test passes only if:

- queued tasks from a goroutine execute in order
- `Promise` callbacks are flushed by the Go ticker
- JS state reaches the expected final snapshot after the queued work completes

## CI Matrix

GitHub Actions workflow: `.github/workflows/v8go-spike.yml`

- Linux and macOS run `go mod tidy`, `go run .`, the event-loop spike, and the
  cold-start benchmarks with archived output files.
- Windows runs `go mod tidy`, captures `go build -x .` output, and records the
  current linker failure as evidence while still running the `goja` baseline
  benchmark with archived output.
- Result interpretation and exit criteria for Fase 1A live in `RESULTS.md`.

## Notes

- This spike is intentionally isolated from the main `go.mod`.
- It uses `github.com/the-btfash-foundation/v8go` because it is more recent than
  the last tagged `rogchap.com/v8go` release and provides a clearer path for
  further evaluation.
- Current observed result on this repository's Windows environment:
  - `go mod tidy` succeeds
  - `go build -x .` reaches the external linker
  - the final link step fails via `g++/ld`
- A dedicated workflow and benchmark harness now exist so future commits can
  collect Linux/macOS evidence without touching the production runtime.
- A dedicated event-loop test harness now exists, but its first real result
  still depends on Linux/macOS CI because the local Windows build remains blocked.
- Cold-start output is now collected with a fixed methodology (`-cpu 1 -count 10`)
  and uploaded as workflow artifacts for later comparison.
