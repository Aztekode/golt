# Golt Roadmap

This roadmap is the public execution view of the longer-term architectural plan
described in `PLAN.md`.

## Current Focus

### 1. Open Source and Governance

- Publish governance and contribution documents
- Improve release process and public project hygiene
- Make the project easier to evaluate and contribute to

### 2. Engine Strategy

- Isolate the current `goja` runtime behind an engine abstraction
- Run a viability spike for `v8go` (or a validated alternative)
- Make the engine decision based on evidence, not assumption

Status:

- `JSEngine` abstraction and `GojaEngine` wrapper are implemented
- `--engine=goja` is available now
- `--engine=v8go` is reserved as the next experimental milestone
- ADR-001 records a short-term no-go for `v8go` as a production Windows engine
- A reproducible smoke test exists in `spikes/v8go-smoke`

### 3. Module and Package Support

- Introduce a first `golt install`
- Add local package cache and lockfile support
- Expand bundling beyond a single entrypoint workflow

## Next Stages

### Runtime and Web APIs

- Better Web-standard APIs
- Better TypeScript diagnostics and runtime messages
- More complete async and streaming behavior

### Data and Security

- Stronger database support
- Permission model for fs, env, and network access

### Tooling and Benchmarks

- Test runner, formatter, linting, type checking
- Public benchmarks and regression tracking

## Execution Order

1. Module 8: Governance and community foundation
2. Module 1: Engine abstraction and viability spike
3. Module 2: Package system
4. Module 5 + Module 10: Security and DX improvements
5. Module 4: Database expansion
6. Module 6: Development toolchain
7. Module 7: Benchmarks
8. Module 9: Deployment/platform support
