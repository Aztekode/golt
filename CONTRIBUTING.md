# Contributing to Golt

Thanks for your interest in contributing to Golt.

## Principles

- Keep changes focused and small.
- Prefer incremental improvements over large rewrites.
- Preserve backward compatibility when possible.
- Document behavior changes in `README.md` and `CHANGELOG.md`.

## Development Setup

From the repository root:

```bash
go test ./...
go run ./cmd/golt --version
```

If you are working on the installer or release flow, also review:

- `.github/workflows/release.yml`
- `installer.iss`

## Pull Requests

Before opening a pull request:

1. Make sure the branch builds with `go test ./...`
2. Update docs if the user-facing behavior changes
3. Add or update examples when it improves discoverability
4. Keep commits readable and scoped to one concern

## Reporting Bugs

Please include:

- Golt version
- Operating system
- Reproduction steps
- Expected behavior
- Actual behavior
- Minimal example, if possible

## Proposing Large Changes

For architectural or multi-module work, open an issue first and reference:

- the relevant section in `PLAN.md`
- the public execution order in `ROADMAP.md`

This helps keep large efforts aligned with the project direction.
