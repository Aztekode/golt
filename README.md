# <p align="center"><img src="docs/images/icon.jpg" alt="Golt icon" width="96" height="96" /></p>

# Golt Runtime

Golt is a lightweight TypeScript/JavaScript backend runtime written in Go. It bundles a TypeScript entry file with esbuild and executes it in a Go-based JavaScript engine using `goja` + `goja_nodejs` event loop.

Golt is intentionally small: instead of trying to clone Node.js, it exposes a focused set of backend primitives through explicit runtime globals.

- CLI: `golt init <project-name>`, `golt run <file.ts>`, `golt watch <file.ts>`
- Runtime globals: `console.log`, `Golt.env`, `Golt.logger`, `Golt.App`, `Golt.db`, `Golt.fs`, `Golt.crypto`, `Golt.jwt`, and global `fetch()`
- Editor support: [Golt VS Code Extension](https://marketplace.visualstudio.com/items?itemName=Aztekode.golt-vscode)

## Project Overview

Golt focuses on a simple workflow:

1. Create a project with `golt init <project-name>`.
2. Write TypeScript.
3. Run it with `golt run app.ts`.
4. Use explicit runtime primitives for HTTP, database access, filesystem, crypto/JWT, and outbound HTTP.

This repo contains:

- A CLI in `cmd/golt`
- The runtime engine and native modules in `runtime`
- A GitHub Pages documentation site in `docs`

## Requirements

- Go (per `go.mod`): Go 1.25.7+

## Installation

### Install from source

From the repo root:

```bash
go build -o golt.exe .\cmd\golt
```

Or install into your Go bin:

```bash
go install .\cmd\golt
```

### Install via `go install` from GitHub

```bash
go install github.com/atrox39/golt/cmd/golt@latest
```

## Quick Start

Initialize a new Golt project:

```bash
golt init my-api
```

This creates:

```txt
my-api/
  app.ts
  golt.json
  .vscode/
    extensions.json
```

The generated `golt.json` marks the folder as a Golt workspace. When opened in VS Code, the Golt extension can detect it and prepare the TypeScript typings automatically.

```bash
cd my-api
code .
golt run app.ts
```

Windows example:

```bash
golt run .\app.ts
```

Watch mode:

```bash
golt watch .\app.ts
```

## Generated Project

### `app.ts`

```ts
const app = Golt.App();

app.use(Golt.logger({ format: "dev" }));

app.get("/", (ctx) => {
  ctx.Json({
    message: "Hello from Golt!",
    runtime: "golt",
  });
});

app.serve(3000);
```

### `golt.json`

```json
{
  "name": "my-api",
  "description": "A Golt Runtime project",
  "version": "0.1.0"
}
```

### `.vscode/extensions.json`

```json
{
  "recommendations": [
    "Aztekode.golt-vscode"
  ]
}
```

## TypeScript Typings

Golt no longer relies on `golt init` to generate `golt.d.ts` directly in the project root.

Typings are handled by the VS Code extension. The extension activates when it detects `golt.json`, copies the bundled Golt typings into `.golt/types/golt/index.d.ts`, and creates or updates `tsconfig.json` / `jsconfig.json` with the required `typeRoots` and `include` entries.

This keeps the CLI scaffold cleaner and allows the editor integration to own typing updates.

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

### HTTP server

Golt exposes an HTTP server through `Golt.App()`, which provides:

- Routing: `app.get/post/put/delete(path, handler)`
- Middleware: `app.use(middleware)`
- Static files: `app.static(prefix, dirPath, spa?)`
- Not found handler: `app.notFound(handler)`
- Server start: `app.serve(port)`

Route patterns use Go's `net/http` `ServeMux` patterns, including path parameters like `/users/{id}`. Access params via `ctx.Param("id")`.

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

  ctx.Status(201).Json({
    ok: true,
    name: body.name,
  });
});

app.static("/public", "./public", true);

app.notFound((ctx) => {
  ctx.Status(404).Send("not found");
});

app.serve(3000);
```

If a route handler or middleware chain finishes without sending a response, Golt finalizes the request automatically with `204 No Content` instead of leaving the HTTP request hanging.

### Outbound HTTP with `fetch`

```ts
fetch("https://httpbin.org/json")
  .then((res) => res.json())
  .then((data) => console.log(data));
```

### Database with `Golt.db`

`query()` is intended for statements that return rows. Use `exec()` for statements that modify state or schema.

```ts
const db = Golt.db.connect("sqlite", "./app.db");

await db.exec(`
  CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
  )
`);

const insert = await db.exec(
  "INSERT INTO users(name) VALUES (?)",
  "Ada"
);

console.log("Insert result:", insert);

const users = await db.query<{ id: number; name: string }>(
  "SELECT id, name FROM users"
);

console.log(users);

db.close();
```

### File system

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

### `golt init <project-name>`

Creates a new Golt project folder:

```txt
<project-name>/
  app.ts
  golt.json
  .vscode/
    extensions.json
```

Notes:

- The command expects exactly one project name.
- It does not generate `golt.d.ts` or `tsconfig.json`.
- Typings are handled by the Golt VS Code extension when the workspace contains `golt.json`.

### `golt run <filename.ts>`

Bundles the entry file and runs it in the Golt runtime.

Notes:

- `golt run` expects exactly one filename argument.
- The bundle is produced with esbuild (`Bundle: true`) and executed as an IIFE.
- Runtime modules are registered before the compiled script is executed.

### `golt watch <filename.ts>`

Runs the file and automatically restarts the process when `.ts` / `.js` files change.

## Runtime API

### `console.log(...args: any[])`

Prints values to stdout.

### `Golt.env: Record<string, string | undefined>`

A map of environment variables read from the host process.

### `fetch(url: string, options?): Promise<FetchResponse>`

Minimal `fetch()` implementation using Go's `net/http`.

- `options.method?: "GET" | "POST" | "PUT" | "DELETE" | "PATCH"`
- `options.headers?: Record<string, string>`
- `options.body?: string`
- `options.timeout?: number` in milliseconds, default `15000`

### `Golt.App(): AppInstance`

Creates an HTTP application instance with its own `ServeMux`.

Behavior notes:

- Uses a per-app `http.NewServeMux()`.
- Registers method-aware routes using patterns like `"GET /users/{id}"`.
- Internally, routes and middleware are executed on the Golt event loop (`RunOnLoop`) to keep JS execution serialized.
- Middleware runs before the route handler through a simple `next()` chain.
- Handler errors log `Error HTTP:` and respond with HTTP 500.
- If a handler finishes without responding, Golt automatically responds with `204 No Content`.
- `app.serve(port)` starts the server and installs SIGINT/SIGTERM shutdown handling with graceful shutdown.

#### `AppInstance`

- `app.use(middleware)` — registers middleware. Middleware signature is `(ctx, next) => void`.
- `app.get/post/put/delete(path, handler)` — registers a handler for a method + path pattern.
- `app.static(prefix, dirPath, spa?)` — serves files from `dirPath` under `prefix`. When `spa` is `true`, missing files fall back to `dirPath/index.html`.
- `app.notFound(handler)` — registers a fallback handler for unmatched routes.
- `app.serve(port)` — starts the HTTP server on `:<port>`.

### `Golt.logger(config?): Middleware`

Creates an HTTP logging middleware compatible with `app.use(...)`.

Config:

- `format: "dev" | "tiny" | "json"`; default `"dev"`

### `Context`

Handlers and middleware receive a context object with the following methods:

- `ctx.Method(): string` — returns the request method.
- `ctx.Url(): string` — returns the request path.
- `ctx.Param(name: string): string` — returns a path parameter.
- `ctx.GetHeader(key: string): string` — reads a request header.
- `ctx.SetHeader(key: string, value: string): void` — sets a response header.
- `ctx.Query(key: string): string` — reads a query string value.
- `ctx.Set(key: string, value: any): void` — sets a per-request local value.
- `ctx.Get<T = any>(key: string): T | undefined` — gets a per-request local value.
- `ctx.Status(code: number): Context` — sets a status code and returns the context.
- `ctx.Send(body: string): void` — writes the status and response body.
- `ctx.Json(data: any): void` — sets `Content-Type: application/json` and writes JSON.
- `ctx.ValidateBody(schema): object | null` — validates JSON request body against a simple schema.

### `Golt.db`

Database access via Go `database/sql`.

- `Golt.db.connect(dialect, dsn): DatabaseClient` — creates a connection and returns a client.
- `Golt.db.query(sql, ...args): Promise<any[]>` — runs a SQL statement that returns rows.
- `Golt.db.exec(sql, ...args): Promise<ExecResult>` — runs a SQL statement that does not return rows.

`ExecResult`:

```ts
interface ExecResult {
  rowsAffected: number | null;
  lastInsertId: number | null;
}
```

Dialects:

- `"sqlite"` using `modernc.org/sqlite`
- `"postgres"` using `lib/pq`
- `"mysql"` using `go-sql-driver/mysql`
- `"sqlserver"` using `go-mssqldb`

### `Golt.fs`

Minimal filesystem helpers.

- `Golt.fs.readFile(path): string`
- `Golt.fs.writeFile(path, content): void`

### `Golt.crypto`

Password hashing utilities using bcrypt.

- `Golt.crypto.hash(password, cost?): Promise<string>`
- `Golt.crypto.compare(password, hash): Promise<boolean>`

### `Golt.jwt`

JWT helpers using HS256.

- `Golt.jwt.sign(payload, secret, expHours?): string`
- `Golt.jwt.verify(token, secret): object | null`

## Repository Layout

- `cmd/golt/main.go`: CLI entrypoint and Cobra commands
- `runtime/engine.go`: compilation with esbuild and execution with goja + event loop
- `runtime/console.go`: `console.log` native module
- `runtime/env.go`: `Golt.env` native module
- `runtime/context.go`: HTTP `Context` implementation
- `runtime/logger.go`: `Golt.logger` middleware factory
- `runtime/http.go`: `Golt.App()` HTTP app
- `runtime/db.go`: `Golt.db` module
- `runtime/fs.go`: `Golt.fs` module
- `runtime/fetch.go`: global `fetch()` implementation
- `runtime/crypto.go`: `Golt.crypto` + `Golt.jwt`

## Documentation Site

A GitHub Pages-ready documentation site is provided at:

- `docs/index.html`

If you enable GitHub Pages, configure Pages to serve from the `docs/` folder.

- Docs entry: `docs/index.html`
- From this README: [Documentation Site](docs/index.html)
- VS Code Extension: [Golt](https://marketplace.visualstudio.com/items?itemName=Aztekode.golt-vscode)

## Contributing

Contributions are welcome.

- Fork the repo and create a feature branch.
- Keep changes focused and include a clear description.
- Add or update examples when you change runtime behavior.
- Run basic checks locally:

```bash
go test ./...
go vet ./...
```

## License

No LICENSE file is currently present in this repository. Until a license is added, the project should be considered proprietary / "all rights reserved" by default.

If you intend this project to be open source, add a `LICENSE` file and update this section to match.
