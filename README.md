# Golt Runtime

Golt is a lightweight TypeScript/JavaScript backend runtime written in Go. It bundles and compiles a TypeScript entry file (via esbuild) and executes it in a Go-based JS engine with a small set of built-in globals.

- CLI: `golt init`, `golt run <file.ts>`
- Runtime globals (today): `console.log`, `Golt.env`, `Golt.App`, `Golt.logger`

## Project Overview

Golt focuses on a simple workflow:

1. Write TypeScript.
2. Run it with `golt run`.
3. Access a small runtime surface area for server-side scripting (starting with logging and environment variables).

This repo contains:
- A CLI in `cmd/golt`
- The runtime engine and native modules in `runtime`

## Installation

### Prerequisites
- Go (per `go.mod`): Go 1.24.3+

### Install (from source)
From the repo root:

```bash
go build -o golt.exe .\cmd\golt
```

Or install into your Go bin:

```bash
go install .\cmd\golt
```

### Install (via `go install` from GitHub)
If the repository is reachable from your environment:

```bash
go install github.com/atrox39/golt/cmd/golt@latest
```

## Quick Start

Initialize a new Golt project in an empty directory:

```bash
golt init
```

This creates:
- `golt.d.ts` (global typings)
- `tsconfig.json`
- `app.ts`

Run the generated entry file:

```bash
golt run app.ts
```

Windows example:

```bash
golt run .\app.ts
```

## Usage Examples

### Logging
```ts
console.log("Hello from Golt");
console.log({ ok: true, now: Date.now() });
```

### Environment variables
Golt exposes environment variables via `Golt.env`:

```ts
console.log("PATH =", Golt.env["PATH"]);
console.log("NODE_ENV =", Golt.env["NODE_ENV"]);
```

### HTTP server (App + routing + middleware)
Golt exposes an HTTP server through `Golt.App()`, which provides:
- Routing: `app.get/post/put/delete(path, handler)`
- Middleware: `app.use(middleware)`
- Not found handler: `app.notFound(handler)`
- Server start: `app.serve(port)`

Route patterns use Go’s `net/http` `ServeMux` patterns (Go 1.22+), including path parameters like `/users/{id}`. Access params via `ctx.Param("id")`.

```ts
const app = Golt.App();

app.use(Golt.logger({ format: "dev" }));

app.get("/", (ctx) => {
  ctx.Send("Hello from Golt HTTP");
});

app.get("/users/{id}", (ctx) => {
  ctx.Json({ id: ctx.Param("id") });
});

app.post("/users", (ctx) => {
  const body = ctx.ValidateBody({ name: "string" });
  if (!body) return;
  ctx.Status(201).Json({ ok: true, name: body.name });
});

app.notFound((ctx) => {
  ctx.Status(404).Send("not found");
});

app.serve(3000);
```

## CLI Reference

### `golt init`
Creates a minimal TypeScript project scaffold for running with Golt:
- `golt.d.ts`
- `tsconfig.json`
- `app.ts`

### `golt run <filename.ts>`
Bundles the entry file and runs it in the Golt runtime.

Notes:
- The runtime currently provides `console.log`, `Golt.env`, `Golt.App`, and `Golt.logger`.
- `golt run` expects a filename argument. Running `golt run` with no filename may crash due to missing argument handling.

## Runtime API (Current)

### `console.log(...args: any[])`
Prints values to stdout.

### `Golt.env: Record<string, string | undefined>`
A map of environment variables read from the host process.

### `Golt.App(): AppInstance`
Creates an HTTP application instance with its own `ServeMux`.

Behavior notes (current implementation):
- Uses a per-app `http.NewServeMux()` and registers method-aware routes using patterns like `"GET /users/{id}"`.
- Internally, routes and middleware are executed on the Golt event loop (`RunOnLoop`) to keep JS execution serialized.
- Middleware runs before the route handler (a simple `next()` chain). Handler errors log `Error HTTP:` and respond with HTTP 500.
- `app.serve(port)` starts the server and installs SIGINT/SIGTERM shutdown handling (graceful shutdown with a 5s timeout).

#### `AppInstance`
- `app.use(middleware)` — registers middleware. Middleware signature is `(ctx, next) => void`.
- `app.get/post/put/delete(path, handler)` — registers a handler for a method + path pattern.
- `app.notFound(handler)` — registers a fallback handler for unmatched routes (status is pre-set to 404).
- `app.serve(port)` — starts the HTTP server on `:<port>`.

### `Golt.logger(config?): Middleware`
Creates an HTTP logging middleware compatible with `app.use(...)`.

Config:
- `format: "dev" | "tiny"` (default `"dev"`)

#### `Context` (HTTP)
Handlers and middleware receive a context object with the following methods:

- `ctx.Method(): string` — returns the request method.
- `ctx.Url(): string` — returns the request path (`r.URL.Path`).
- `ctx.Param(name: string): string` — returns a path parameter (`r.PathValue(name)`), when using patterns like `/users/{id}`.
- `ctx.Status(code: number): Context` — sets a status code (chainable).
- `ctx.Send(body: string): void` — writes the status (defaults to 200) and response body.
- `ctx.Json(data: any): void` — sets `Content-Type: application/json` and writes JSON.
- `ctx.ValidateBody(schema): object | null` — validates JSON request body against a simple schema (returns `null` and responds 400 on invalid input).

Typing is generated by `golt init` in `golt.d.ts`:
```ts
declare namespace Golt {
  export const env: Record<string, string | undefined>;
}
```
`golt init` currently generates a richer `golt.d.ts` that includes `AppInstance`, `Context`, middleware types, and schema inference helpers.

## Repository Layout

- `cmd/golt/main.go`: CLI entrypoint (Cobra commands)
- `runtime/engine.go`: compilation (esbuild) + execution (goja + event loop)
- `runtime/console.go`: `console.log` native module
- `runtime/env.go`: `Golt.env` native module
- `runtime/context.go`: HTTP `Context` implementation (params, JSON, body validation)
- `runtime/logger.go`: `Golt.logger` middleware factory
- `runtime/http.go`: `Golt.App()` HTTP app (routes, middleware, server)

## Documentation Site (GitHub Pages)

A GitHub Pages-ready documentation site is provided at:

- `docs/index.html`

If you enable GitHub Pages, configure Pages to serve from the `docs/` folder.

- Docs entry: `docs/index.html`
- From this README: [Documentation Site](docs/index.html)

## Contributing

Contributions are welcome.

- Fork the repo and create a feature branch
- Keep changes focused and include a clear description
- Add/update examples when you change runtime behavior
- Run basic checks locally:
  - `go test ./...` (if/when tests exist)
  - `go vet ./...`

## License

No LICENSE file is currently present in this repository. Until a license is added, the project should be considered proprietary / “all rights reserved” by default.

If you intend this project to be open source, add a `LICENSE` file (for example: MIT, Apache-2.0, or GPL-3.0) and update this section to match.
