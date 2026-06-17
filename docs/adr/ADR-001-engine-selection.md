# ADR-001: Engine Selection Strategy

- Status: Accepted
- Date: 2026-06-15

## Context

Golt currently runs on `goja` and has just introduced a first `JSEngine`
abstraction layer. The next planned step was to validate whether `v8go` is a
viable replacement or secondary engine for future performance-oriented work.

Constraints relevant to this decision:

- Golt already ships a Windows installer and currently targets Windows users.
- The main module must remain stable while research happens.
- A candidate engine must support isolate/context creation, script execution,
  globals, and promise/microtask behavior before deeper API integration.

## Research Inputs

### Local repository facts

- Main runtime uses `goja` in [engine.go](file:///c:/Users/cortega/Development/personal/golt-project/golt/runtime/engine.go).
- The CLI already supports `--engine=goja` and reserves `--engine=v8go`.
- A new isolated spike lives in [spikes/v8go-smoke](file:///c:/Users/cortega/Development/personal/golt-project/golt/spikes/v8go-smoke).

### External signals

- The last tagged upstream `rogchap.com/v8go` release visible on pkg.go.dev is
  `v0.10.0` from 2023, and its changelog explicitly notes that Windows support
  had been removed in earlier releases and remained problematic [pkg.go.dev](https://pkg.go.dev/rogchap.com/v8go@v0.9.0).
- The upstream GitHub repository still shows activity and open upgrade PRs in
  2025-2026, but release consumption remains stale and operational support looks
  uneven [GitHub PRs](https://github.com/rogchap/v8go/pulls).
- More recent forks exist:
  - `github.com/the-btfash-foundation/v8go` published a newer module in October
    2025 [pkg.go.dev](https://pkg.go.dev/github.com/the-btfash-foundation/v8go)
  - `github.com/tommie/v8go` documents active maintenance and V8 upgrades, but
    still states that Windows binary support used to exist and currently needs
    external work [README](https://github.com/tommie/v8go/blob/master/README.md)

## Spike

The spike uses `github.com/the-btfash-foundation/v8go` in a separate module to
avoid contaminating the main `go.mod`.

Files:

- [go.mod](file:///c:/Users/cortega/Development/personal/golt-project/golt/spikes/v8go-smoke/go.mod)
- [main.go](file:///c:/Users/cortega/Development/personal/golt-project/golt/spikes/v8go-smoke/main.go)
- [README.md](file:///c:/Users/cortega/Development/personal/golt-project/golt/spikes/v8go-smoke/README.md)

The smoke test validates:

- isolate/context creation
- global mutation from Go
- script execution
- promise callback + microtask checkpoint

## Observed Result on Current Environment

Environment:

- OS: Windows
- Arch: amd64
- `CGO_ENABLED=1`
- compiler toolchain detected: `gcc` / `g++`

Results:

1. `go mod tidy` in the spike module succeeds.
2. `go build -x .` compiles the Go and C++ bridge code far enough to invoke the
   external linker.
3. The final Windows link step fails through `g++`/`ld` with exit status `1`
   and `collect2.exe: error: ld returned 5 exit status`.
4. A dedicated GitHub Actions workflow now exists to rerun the smoke test on
   Linux x64 and macOS while preserving the Windows linker log as evidence.
5. A small cold-start benchmark harness now exists to compare `goja` against
   `v8go` on platforms where the spike links successfully.
6. A dedicated event-loop harness now exists to verify whether a Go-owned loop
   with `time.Ticker` plus goroutine-fed task queues can flush V8 microtasks
   without touching the production runtime.
7. The cold-start benchmark process is now fixed to `-cpu 1 -count 10`, and the
   raw outputs are archived per platform in CI so later comparisons are based on
   preserved evidence instead of ad-hoc console output.

This confirms that, on the current Windows setup, the candidate `v8go` path is
not yet operational enough to adopt as the next production engine.

## Decision

Short-term decision: **No-go for immediate migration from `goja` to `v8go`.**

Accepted path:

- Keep `goja` as the default and only working engine.
- Preserve the new engine abstraction layer.
- Keep `--engine=v8go` as a reserved experimental target, not a supported mode.
- Continue future engine research through isolated spikes, not through the main
  runtime.

Candidate for future reevaluation:

- Prefer evaluating maintained forks such as
  `github.com/the-btfash-foundation/v8go` or `github.com/tommie/v8go`, not the
  stale tagged upstream alone.

## Consequences

### Positive

- No disruption to the current Windows release/distribution path.
- The architecture still moves forward because the runtime is now abstracted.
- Future engine experiments stay low-risk and reproducible.

### Negative

- Golt does not get a production-ready JIT engine yet.
- Package compatibility and some advanced JS semantics remain bounded by `goja`
  for now.

## Next Actions

1. Repeat the spike on Linux x64 and macOS ARM64 in CI or dedicated machines.
2. Capture the exact Windows linker requirements if a fork becomes promising.
3. Collect the first benchmark outputs from the new harness on Linux/macOS and
   decide whether deeper event-loop work is justified.
4. Collect the first event-loop spike result from Linux/macOS and confirm
   whether the hybrid loop model behaves predictably under CI.
5. Collect the first archived cold-start benchmark outputs from Linux/macOS and
   compare them against the Windows `goja` baseline.
6. Benchmark `goja` vs candidate engines more broadly only after a
   cross-platform build path exists.
7. Consider QuickJS-based alternatives if V8 remains too costly operationally.
