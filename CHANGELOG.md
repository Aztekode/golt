# Changelog

This project follows a simple, release-oriented changelog. Roadmap and
governance changes are tracked here when they materially affect contributors or
project adoption.

## v1.0.3

- Added update detection (local version change) and remote update checks (GitHub releases) with cache.
- Added release verification using `SHA256SUMS` + `SHA256SUMS.minisig` and a new `golt verify-release` command.
- Improved TypeScript compilation diagnostics and enabled inline sourcemaps for better runtime errors.
- Fixed HTTP async handlers so requests don't finish before awaited promises resolve.
- Added `examples/` folder and made it optionally installable via the Windows installer.
- Improved Windows installer (Inno Setup): per-user/per-machine choice, branding assets, optional examples, PATH cleanup on uninstall.
- Added initial open source governance documents: LICENSE, CONTRIBUTING, CODE_OF_CONDUCT, SECURITY, ROADMAP, and GitHub collaboration templates.
- Added the first engine abstraction layer with `JSEngine`, a concrete `GojaEngine` wrapper, and CLI engine selection via `--engine`.
- Added an isolated `v8go` smoke test and ADR-001 documenting a short-term no-go for production migration on Windows.
