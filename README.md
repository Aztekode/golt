# <p align="center"><img src="docs/images/icon.jpg" alt="Golt icon" width="96" height="96" /></p>

# Golt Runtime

Golt is a lightweight TypeScript/JavaScript backend runtime written in Go. It bundles a TypeScript entry file (via esbuild) and executes it in a Go-based JS engine (goja + goja_nodejs event loop), with a small set of built-in globals focused on backend scripting.

- CLI: `golt init`, `golt run <file.ts>`, `golt watch <file.ts>`
- Runtime globals (current): `console.log`, `Golt.env`, `Golt.logger`, `Golt.App`, `Golt.db`, `Golt.fs`, `Golt.crypto`, `Golt.jwt`, and global `fetch()`

## Project Overview

Golt focuses on a simple workflow:

1. Write TypeScript.
2. Run it with `golt run`.
3. Use explicit runtime primitives for HTTP, database access, filesystem, crypto/JWT, and outbound HTTP.

This repo contains:
- A CLI in `cmd/golt`
- The runtime engine and native modules in `runtime`

## Requirements

- Go (per `go.mod`): Go 1.25.7+

## Installation

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

Watch mode (auto-restart on `.ts` / `.js` changes):

```bash
golt watch .\app.ts
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

### HTTP server (App + routing + middleware + static)
Golt exposes an HTTP server through `Golt.App()`, which provides:
- Routing: `app.get/post/put/delete(path, handler)`
- Middleware: `app.use(middleware)`
- Static files: `app.static(prefix, dirPath, spa?)`
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

app.static("/public", "./public", true);

app.notFound((ctx) => {
  ctx.Status(404).Send("not found");
});

app.serve(3000);
```

### Outbound HTTP (fetch)

```ts
fetch("https://httpbin.org/json")
  .then((res) => res.json())
  .then((data) => console.log(data));
```

### Database (Golt.db)

```ts
const db = Golt.db.connect("sqlite", "./app.db");

db.query("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT)");
db.query("INSERT INTO users(name) VALUES (?)", "Ada");

db.query("SELECT id, name FROM users").then((rows) => console.log(rows));
```

### File system (Golt.fs)

```ts
Golt.fs.writeFile("./hello.txt", "Hello from Golt\n");
console.log(Golt.fs.readFile("./hello.txt"));
```

### Crypto + JWT

```ts
Golt.crypto.hash("password123").then((hash) => {
  console.log("hash =", hash);
  return Golt.crypto.compare("password123", hash);
}).then((ok) => console.log("match =", ok));

const token = Golt.jwt.sign({ sub: "user_123" }, "secret", 24);
const payload = Golt.jwt.verify(token, "secret");
console.log(payload);
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
- `golt run` expects exactly one filename argument.
- The bundle is produced with esbuild (`Bundle: true`) and executed as an IIFE.

### `golt watch <filename.ts>`
Runs the file and automatically restarts the process when `.ts` / `.js` files change in the current directory tree.

## Runtime API (Current)

### `console.log(...args: any[])`
Prints values to stdout.

### `Golt.env: Record<string, string | undefined>`
A map of environment variables read from the host process.

### `fetch(url: string, options?): Promise<FetchResponse>`
Minimal `fetch()` implementation using Go’s `net/http`.

- `options.method?: "GET" | "POST" | "PUT" | "DELETE" | "PATCH"`
- `options.headers?: Record<string, string>`
- `options.body?: string`
- `options.timeout?: number` (milliseconds, default 15000)

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
- `app.static(prefix, dirPath, spa?)` — serves files from `dirPath` under `prefix`. When `spa` is `true`, missing files fall back to `dirPath/index.html`.
- `app.notFound(handler)` — registers a fallback handler for unmatched routes (status is pre-set to 404).
- `app.serve(port)` — starts the HTTP server on `:<port>`.

### `Golt.logger(config?): Middleware`
Creates an HTTP logging middleware compatible with `app.use(...)`.

Config:
- `format: "dev" | "tiny" | "json"` (default `"dev"`)

#### `Context` (HTTP)
Handlers and middleware receive a context object with the following methods:

- `ctx.Method(): string` — returns the request method.
- `ctx.Url(): string` — returns the request path (`r.URL.Path`).
- `ctx.Param(name: string): string` — returns a path parameter (`r.PathValue(name)`), when using patterns like `/users/{id}`.
- `ctx.GetHeader(key: string): string` — reads a request header.
- `ctx.SetHeader(key: string, value: string): void` — sets a response header.
- `ctx.Query(key: string): string` — reads a query string value.
- `ctx.Set(key: string, value: any): void` — sets a per-request local value.
- `ctx.Get<T = any>(key: string): T | undefined` — gets a per-request local value.
- `ctx.Status(code: number): Context` — sets a status code (chainable).
- `ctx.Send(body: string): void` — writes the status (defaults to 200) and response body.
- `ctx.Json(data: any): void` — sets `Content-Type: application/json` and writes JSON.
- `ctx.ValidateBody(schema): object | null` — validates JSON request body against a simple schema (returns `null` and responds 400 on invalid input).

### `Golt.db`
Database access via Go `database/sql`.

- `Golt.db.connect(dialect, dsn): DatabaseClient` — creates a connection and returns a client.
- `Golt.db.query(sql, ...args): Promise<any[]>` — runs a query on the active connection.

Dialects:
- `"sqlite"` (driver: `modernc.org/sqlite`)
- `"postgres"` (driver: `lib/pq`)
- `"mysql"` (driver: `go-sql-driver/mysql`)
- `"sqlserver"` (driver: `go-mssqldb`)

### `Golt.fs`
Minimal filesystem helpers.

- `Golt.fs.readFile(path): string`
- `Golt.fs.writeFile(path, content): void`

### `Golt.crypto`
Password hashing utilities (bcrypt).

- `Golt.crypto.hash(password, cost?): Promise<string>`
- `Golt.crypto.compare(password, hash): Promise<boolean>`

### `Golt.jwt`
JWT helpers (HS256).

- `Golt.jwt.sign(payload, secret, expHours?): string`
- `Golt.jwt.verify(token, secret): object | null`

Typing is generated by `golt init` in `golt.d.ts`, including `fetch()` and all `Golt.*` APIs.

## Repository Layout

- `cmd/golt/main.go`: CLI entrypoint (Cobra commands)
- `runtime/engine.go`: compilation (esbuild) + execution (goja + event loop)
- `runtime/console.go`: `console.log` native module
- `runtime/env.go`: `Golt.env` native module
- `runtime/context.go`: HTTP `Context` implementation (params, JSON, body validation)
- `runtime/logger.go`: `Golt.logger` middleware factory
- `runtime/http.go`: `Golt.App()` HTTP app (routes, middleware, server)
- `runtime/db.go`: `Golt.db` module (database/sql)
- `runtime/fs.go`: `Golt.fs` module (read/write)
- `runtime/fetch.go`: global `fetch()` implementation
- `runtime/crypto.go`: `Golt.crypto` + `Golt.jwt`

## Documentation Site (GitHub Pages)

A GitHub Pages-ready documentation site is provided at:

- `docs/index.html`

If you enable GitHub Pages, configure Pages to serve from the `docs/` folder.

- Docs entry: `docs/index.html`
- From this README: [Documentation Site](docs/index.html)
- VS Code Extension: https://marketplace.visualstudio.com/items?itemName=Aztekode.golt-vscode

## Contributing

Contributions are welcome.

- Fork the repo and create a feature branch
- Keep changes focused and include a clear description
- Add/update examples when you change runtime behavior
- Run basic checks locally:
  - `go test ./...`
  - `go vet ./...`

## License

No LICENSE file is currently present in this repository. Until a license is added, the project should be considered proprietary / “all rights reserved” by default.

If you intend this project to be open source, add a `LICENSE` file (for example: MIT, Apache-2.0, or GPL-3.0) and update this section to match.
