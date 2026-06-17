# Golt Runtime — Plan de Modernización Profesional
## Documento de Análisis y Corrección por Módulos para Agentes de IA

> **Versión del documento:** 1.0  
> **Fecha:** 2026-06-15  
> **Objetivo:** Transformar Golt de un runtime de scripting experimental en una alternativa profesional a Node.js/Deno/Bun, trabajando por módulos independientes para evitar carga excesiva.  
> **Motor actual:** goja (intérprete JS puro en Go)  
> **Motor propuesto:** v8go (bindings de V8 para Go) — ver Módulo 1.

---

## Índice de Módulos

| # | Módulo | Prioridad | Complejidad | Bloqueante para producción |
|---|--------|-----------|-------------|---------------------------|
| 1 | Migración de Motor: goja → v8go | 🔴 Crítica | Alta | Sí |
| 2 | Sistema de Módulos y Compatibilidad npm | 🔴 Crítica | Muy Alta | Sí |
| 3 | API Surface y Web Standards | 🟡 Alta | Media | Parcial |
| 4 | Drivers de Base de Datos | 🟡 Alta | Media | Sí |
| 5 | Seguridad y Modelo de Permisos | 🟡 Alta | Media | No (pero enterprise) |
| 6 | Toolchain de Desarrollo (Test, Debug, Lint) | 🟢 Media | Media | No |
| 7 | Rendimiento y Benchmarking | 🟢 Media | Baja | No |
| 8 | Gobernanza, Open Source y Comunidad | 🔴 Crítica | Baja | Sí |
| 9 | Deployment y Plataformas | 🟢 Media | Baja | No |
| 10 | Documentación y DX (Developer Experience) | 🟢 Media | Baja | No |

---

## Instrucciones para el Agente de IA

Este documento está diseñado para ser procesado **módulo por módulo**. Cada módulo es independiente salvo donde se indiquen dependencias explícitas. El agente debe:

1. **Analizar** el estado actual de Golt respecto al módulo.
2. **Identificar** gaps y riesgos técnicos.
3. **Proponer** un plan de implementación con milestones.
4. **Generar** código, configuraciones o documentación según corresponda.
5. **Validar** contra benchmarks de Node.js 24, Deno 2.6 y Bun 1.3.

---

## Módulo 1: Migración de Motor — goja → v8go

### 1.1 Estado Actual
- **Motor:** goja (github.com/dop251/goja)
- **Tipo:** Intérprete JavaScript 100% Go, sin JIT.
- **Problemas:**
  - Rendimiento CPU-bound 10x–50x inferior a V8.
  - No soporta WebAssembly.
  - No tiene generación de código optimizada.
  - Heap gestionado por Go GC, no por V8 GC.

### 1.2 Propuesta: v8go
- **Repositorio:** https://github.com/rogchap/v8go
- **Descripción:** Bindings de Go para el motor V8 de Google (el mismo de Node.js y Deno).
- **Ventajas:**
  - JIT compilation (TurboFan).
  - Compatibilidad casi total con ECMAScript 2024.
  - WebAssembly nativo.
  - Inspector protocol para debugging.
  - Mejor gestión de memoria para cargas JS intensivas.
- **Desventajas/Riesgos:**
  - Dependencia de C++ (V8) → compilación más compleja, cross-compilation difícil.
  - Tamaño de binario aumenta (~15–25 MB).
  - Requiere CGO (goja es puro Go).
  - Menos control sobre el event loop (hay que integrarlo manualmente).

### 1.3 Plan de Migración por Fases

#### Fase 1A: Spike de viabilidad (1–2 semanas)
- [ ] Compilar v8go en Linux x64, macOS ARM64, Windows x64.
- [x] Crear un `main.go` mínimo que ejecute un script JS simple.
- [ ] Medir cold start vs goja.
- [ ] Verificar que el event loop de Go (`time.Ticker`, goroutines) puede alimentar el event loop de V8.

Notas:

- Existe spike reproducible en `spikes/v8go-smoke`.
- La compilación local en Windows amd64 llega al enlace final pero falla en `g++/ld`; ver `docs/adr/ADR-001-engine-selection.md`.
- Existe workflow aislado `.github/workflows/v8go-spike.yml` para recolectar evidencia en Linux/macOS y conservar el log de Windows.
- Existe harness mínimo de benchmark en `spikes/v8go-smoke/bench` para medir cold start de `goja` y, cuando enlace, de `v8go`.
- Existe subspike de event loop en `spikes/v8go-smoke/eventloop` para validar el patrón `time.Ticker` + goroutines + `PerformMicrotaskCheckpoint()`.
- La medición de cold start está normalizada con `-cpu 1 -count 10` y guarda artefactos por plataforma para comparación posterior.

#### Fase 1B: Abstracción del Motor (2–3 semanas)
- [x] Definir interfaz `JSEngine` en Go:
  ```go
  type JSEngine interface {
      Eval(script string) (Value, error)
      Call(fn string, args ...Value) (Value, error)
      SetGlobal(name string, val Value) error
      RunEventLoop() error
      Dispose()
  }
  ```
- [x] Implementar `GojaEngine` (wrapper actual).
- [ ] Implementar `V8GoEngine` (nuevo).
- [x] Feature flag `--engine=v8go` / `--engine=goja`.

#### Fase 1C: Integración con APIs Nativas (3–4 semanas)
- [ ] Exponer `Golt.App`, `Golt.db`, `Golt.fs` como objetos globales en V8.
- [ ] Implementar `fetch()` usando `net/http` de Go + Promises en V8.
- [ ] Asegurar que `async/await` funcione correctamente en el event loop híbrido.

#### Fase 1D: Benchmarking y Decisión (1 semana)
- [ ] Benchmark HTTP throughput (wrk/wrk2).
- [ ] Benchmark JSON parsing y crypto.
- [ ] Benchmark memoria y latencia p99.
- [ ] Decisión go/no-go para deprecar goja.

Nota:

- Existe una decisión preliminar de corto plazo: no migrar todavía a `v8go` como engine productivo en Windows hasta resolver la viabilidad multiplataforma. Ver ADR-001.

### 1.4 Código de Referencia: Event Loop Híbrido V8+Go

```go
// engine/v8go_eventloop.go
package engine

import (
    "time"
    "github.com/rogchap/v8go"
)

type V8EventLoop struct {
    isolate   *v8go.Isolate
    context   *v8go.Context
    taskQueue chan func()
}

func NewV8EventLoop() *V8EventLoop {
    iso := v8go.NewIsolate()
    ctx := v8go.NewContext(iso)
    return &V8EventLoop{
        isolate:   iso,
        context:   ctx,
        taskQueue: make(chan func(), 1024),
    }
}

func (el *V8EventLoop) Run() {
    ticker := time.NewTicker(10 * time.Millisecond)
    defer ticker.Stop()
    for {
        select {
        case task := <-el.taskQueue:
            task()
        case <-ticker.C:
            // Procesar microtasks de V8
            el.context.PerformMicrotaskCheckpoint()
        }
    }
}

func (el *V8EventLoop) Enqueue(task func()) {
    el.taskQueue <- task
}
```

### 1.5 Dependencias
- **Bloquea:** Módulo 3 (API Surface) — si cambiamos el motor, las APIs nativas deben reimplementarse.
- **Bloquea:** Módulo 7 (Benchmarking) — sin V8 no hay comparación justa.

---

## Módulo 2: Sistema de Módulos y Compatibilidad npm

### 2.1 Estado Actual
- Sin sistema de módulos externo.
- Solo entry point único (`app.ts`) bundlado con esbuild.
- No hay `node_modules`, `package.json`, ni resolución de dependencias.

### 2.2 Objetivo
- Permitir `import { something } from "npm:package"` o `import { something } from "./node_modules/package"`.
- Soporte mínimo para ESM y CommonJS (al menos las top 100 librerías de npm).

### 2.3 Plan por Fases

#### Fase 2A: Cache de Paquetes npm (2 semanas)
- [ ] Implementar `golt install` que descargue tarballs de npm registry.
- [ ] Cache local en `~/.golt/cache/npm/`.
- [ ] Resolución de semver básica.

#### Fase 2B: Bundler con esbuild + npm (3 semanas)
- [ ] Configurar esbuild para resolver desde `node_modules`.
- [ ] Generar `golt.lock` (lockfile) para reproducibilidad.
- [ ] Soporte para `package.json` con campo `"golt"` para configuración específica.

#### Fase 2C: Polyfills de Node.js (4–6 semanas)
- [ ] Mapear APIs de Node.js a APIs de Golt:
  - `fs` → `Golt.fs` (extender API actual)
  - `path` → implementar en Go
  - `crypto` → extender `Golt.crypto` con más algoritmos
  - `http` → `Golt.App` o `fetch`
  - `events` → implementar EventEmitter
  - `stream` → implementar Web Streams o streams básicos
  - `process.env` → `Golt.env`
- [ ] Crear paquete `@golt/polyfill-node` auto-inyectado.

#### Fase 2D: Testing de Compatibilidad (continuo)
- [ ] Crear test suite con las top 50 librerías npm (lodash, axios, express-compat, etc.).
- [ ] Reportar porcentaje de compatibilidad mensual.

### 2.4 Dependencias
- **Requiere:** Módulo 1 (V8) — muchas librerías npm usan features de JS que goja no soporta bien (generators, proxies complejos, etc.).
- **Bloquea:** Todo el ecosistema. Sin esto, Golt es solo un toy runtime.

---

## Módulo 3: API Surface y Web Standards

### 3.1 Estado Actual
- API propietaria: `Golt.App`, `Golt.db`, `Golt.fs`, `Golt.crypto`, `Golt.jwt`, `Golt.env`.
- `fetch()` minimal basado en Go `net/http`.
- Sin Web Streams, Web Crypto, WebSocket, URLPattern, EventTarget.

### 3.2 Objetivo
- Implementar APIs web estándar (W3C/WhatWG) para que el código sea portable entre navegador y servidor.
- Mantener APIs de Golt como extensiones, no como reemplazos.

### 3.3 Plan por Fases

#### Fase 3A: Web Crypto (1 semana)
- [ ] Implementar `crypto.subtle.digest()`, `encrypt()`, `decrypt()`, `sign()`, `verify()`.
- [ ] Usar librerías Go `crypto/sha256`, `crypto/aes`, `crypto/rsa`.
- [ ] Mapear a API Web Crypto estándar.

#### Fase 3B: Web Streams (2 semanas)
- [ ] Implementar `ReadableStream`, `WritableStream`, `TransformStream`.
- [ ] Integrar con `fetch()` para streaming de respuestas.
- [ ] Base para futuro soporte de `fs.createReadStream`.

#### Fase 3C: WebSocket (2 semanas)
- [ ] Implementar `WebSocket` client y server.
- [ ] Server: integrar con `Golt.App` o middleware.
- [ ] Client: usar `gorilla/websocket` desde Go.

#### Fase 3D: EventTarget y AbortController (1 semana)
- [ ] Implementar `EventTarget`, `Event`, `CustomEvent`.
- [ ] Implementar `AbortController` / `AbortSignal` para `fetch()`.

#### Fase 3E: URLPattern (opcional, 1 semana)
- [ ] Implementar `URLPattern` para routing avanzado.

### 3.4 Código de Referencia: Web Crypto en Go

```go
// runtime/webcrypto.go
package runtime

import (
    "crypto/sha256"
    "encoding/base64"
)

func WebCryptoDigest(algorithm string, data []byte) (string, error) {
    switch algorithm {
    case "SHA-256":
        sum := sha256.Sum256(data)
        return base64.StdEncoding.EncodeToString(sum[:]), nil
    default:
        return "", fmt.Errorf("unsupported algorithm: %s", algorithm)
    }
}
```

### 3.5 Dependencias
- **Requiere:** Módulo 1 (V8) — algunas APIs web usan features avanzadas de JS.
- **Requiere:** Módulo 2 (npm compat) — para que librerías que dependen de `stream` o `crypto` funcionen.

---

## Módulo 4: Drivers de Base de Datos

### 4.1 Estado Actual
- Solo SQLite via `database/sql` de Go.
- Métodos: `db.exec()` y `db.query()`.
- Sin pooling, sin prepared statements expuestos, sin transacciones.

### 4.2 Objetivo
- Soporte nativo para PostgreSQL, MySQL, Redis.
- Pooling de conexiones, transacciones, prepared statements.
- Compatibilidad con ORMs (Prisma, Drizzle) a largo plazo.

### 4.3 Plan por Fases

#### Fase 4A: PostgreSQL (2 semanas)
- [ ] Integrar `lib/pq` o `jackc/pgx` en Go.
- [ ] Exponer `Golt.db.connect("postgres", connString)`.
- [ ] Soporte para `db.transaction()`, `db.prepare()`.

#### Fase 4B: MySQL (1 semana)
- [ ] Integrar `go-sql-driver/mysql`.
- [ ] Misma API unificada que PostgreSQL.

#### Fase 4C: Redis (1 semana)
- [ ] Integrar `redis/go-redis`.
- [ ] API: `Golt.redis.connect()`, `client.get()`, `client.set()`, `client.publish()`, etc.

#### Fase 4D: Connection Pooling y Transacciones (2 semanas)
- [ ] Exponer configuración de pool size, timeout, max connections.
- [ ] Implementar `db.beginTransaction()`, `tx.commit()`, `tx.rollback()`.

### 4.4 Dependencias
- **Requiere:** Módulo 1 (V8) — algunos drivers de DB en npm usan código nativo.
- **Requiere:** Módulo 2 (npm compat) — para ORMs.

---

## Módulo 5: Seguridad y Modelo de Permisos

### 5.1 Estado Actual
- Sin sandbox. Todo el código tiene acceso total al sistema de archivos, red y variables de entorno.
- Sin auditoría de supply chain (no hay npm, no hay supply chain).

### 5.2 Objetivo
- Implementar modelo de permisos deny-by-default, similar a Deno.
- Sandboxing de filesystem, red, variables de entorno y ejecución de subprocessos.

### 5.3 Plan por Fases

#### Fase 5A: Flags de Permisos (1 semana)
- [ ] `--allow-read=/path` / `--deny-read`
- [ ] `--allow-write=/path` / `--deny-write`
- [ ] `--allow-net=host:port` / `--deny-net`
- [ ] `--allow-env=VAR1,VAR2` / `--deny-env`
- [ ] `--allow-run` (para futuro soporte de subprocessos)

#### Fase 5B: Enforcement en APIs Nativas (2 semanas)
- [ ] Validar permisos en `Golt.fs.readFile`, `writeFile`.
- [ ] Validar permisos en `fetch()` y `Golt.App`.
- [ ] Validar permisos en `Golt.env`.

#### Fase 5C: Supply Chain Security (futuro)
- [ ] Verificación de checksums en `golt.lock`.
- [ ] Integración con `sigstore` para firmas de paquetes.

### 5.4 Dependencias
- **Independiente** — puede implementarse en paralelo con otros módulos.

---

## Módulo 6: Toolchain de Desarrollo

### 6.1 Estado Actual
- Comandos: `golt init`, `golt run`, `golt watch`.
- Sin test runner, sin debugger, sin linter, sin formatter.

### 6.2 Objetivo
- Toolchain completa integrada: test, debug, lint, format, type-check.

### 6.3 Plan por Fases

#### Fase 6A: Test Runner (2 semanas)
- [ ] Comando `golt test`.
- [ ] Soporte para `describe`, `it`, `expect` (API similar a Vitest/Jest).
- [ ] Mocking de `fetch`, `Golt.db`, `Golt.fs`.
- [ ] Coverage reporting básico.

#### Fase 6B: Debugger (3 semanas)
- [ ] Integrar Inspector Protocol de V8 (requiere Módulo 1).
- [ ] Soporte para VS Code via `launch.json`.
- [ ] Breakpoints, step over/into, watch variables.

#### Fase 6C: Linter y Formatter (2 semanas)
- [ ] Integrar `dprint` o `biome` para formatting.
- [ ] Integrar `biome` o reglas custom para linting.
- [ ] Comandos: `golt fmt`, `golt lint`.

#### Fase 6D: Type Checking (2 semanas)
- [ ] Integrar `tsc` o `typescript` como librería embebida.
- [ ] Comando `golt check` para validación de tipos sin ejecución.

### 6.4 Dependencias
- **Requiere:** Módulo 1 (V8) — Inspector Protocol solo disponible en V8.
- **Independiente** en parte — test runner puede funcionar con goja inicialmente.

---

## Módulo 7: Rendimiento y Benchmarking

### 7.1 Estado Actual
- Sin benchmarks públicos.
- Sin comparativa con Node.js, Deno o Bun.

### 7.2 Objetivo
- Establecer un benchmark suite continuo (CI) que compare Golt vs competidores.

### 7.3 Plan por Fases

#### Fase 7A: Benchmark Suite (1 semana)
- [ ] HTTP throughput: `wrk -t12 -c400 -d30s http://localhost:3000`.
- [ ] JSON serialization/deserialization: 1M objetos.
- [ ] Crypto: 10k hashes SHA-256 / bcrypt.
- [ ] File I/O: 10k lecturas/escrituras concurrentes.
- [ ] Cold start: tiempo desde `golt run` hasta primera request.

#### Fase 7B: CI y Reporting (1 semana)
- [ ] GitHub Actions que corra benchmarks en cada PR.
- [ ] Publicar resultados en README o página web.
- [ ] Alertar si hay regresión >10%.

#### Fase 7C: Optimizaciones Guiadas por Datos (continuo)
- [ ] Profilear con `pprof` de Go.
- [ ] Identificar cuellos de botella en el event loop.
- [ ] Optimizar allocaciones de memoria en el bridge Go↔JS.

### 7.4 Dependencias
- **Requiere:** Módulo 1 (V8) — para comparar motor nuevo vs viejo.

---

## Módulo 8: Gobernanza, Open Source y Comunidad

### 8.1 Estado Actual
- Sin repositorio público visible en GitHub.
- Sin licencia clara.
- Sin roadmap público.
- Sin contribuciones externas.

### 8.2 Objetivo
- Abrir el código bajo licencia permisiva (MIT o Apache-2.0).
- Establecer gobernanza transparente.
- Construir comunidad.

### 8.3 Plan por Fases

#### Fase 8A: Publicación del Código (inmediato)
- [ ] Crear repositorio `github.com/aztekode/golt`.
- [x] Añadir LICENSE (MIT recomendado).
- [x] Añadir CONTRIBUTING.md, CODE_OF_CONDUCT.md.
- [x] Añadir SECURITY.md con política de vulnerabilidades.

#### Fase 8B: Documentación Pública (1 semana)
- [x] README con quickstart, API reference, ejemplos.
- [ ] Documentación web en `golt.dev/docs`.
- [x] Changelog semántico (`CHANGELOG.md`).

#### Fase 8C: Comunidad (continuo)
- [ ] Canal de Discord o GitHub Discussions.
- [ ] Issues etiquetadas `good first issue`.
- [ ] Roadmap público en GitHub Projects.

### 8.4 Dependencias
- **Independiente** — pero bloquea adopción profesional.

---

## Módulo 9: Deployment y Plataformas

### 9.1 Estado Actual
- Docker image: `aztekode/golt:1.0.2`.
- Sin soporte nativo Windows.
- Sin runtime para serverless (AWS Lambda, Cloudflare Workers).

### 9.2 Objetivo
- Multiplataforma nativa (Linux, macOS, Windows).
- Runtime para serverless.
- Integración con plataformas de edge.

### 9.3 Plan por Fases

#### Fase 9A: Multiplataforma (2 semanas)
- [ ] Compilación cruzada para Windows (considerar que v8go + CGO complica esto).
- [ ] macOS ARM64 (Apple Silicon).
- [ ] Publicar binarios en GitHub Releases.

#### Fase 9B: AWS Lambda Runtime (2 semanas)
- [ ] Crear custom runtime para AWS Lambda.
- [ ] Handler que reciba eventos API Gateway y los pase a `Golt.App`.

#### Fase 9C: Cloudflare Workers / Vercel Edge (futuro)
- [ ] Investigar compilación a WebAssembly (si se usa v8go, esto es complejo; si se mantiene goja, más factible).
- [ ] O alternativa: ofrecer hosting propio tipo "Golt Deploy".

### 9.4 Dependencias
- **Requiere:** Módulo 1 (V8) — v8go complica cross-compilation; evaluar si se mantiene goja para WASM.

---

## Módulo 10: Documentación y DX (Developer Experience)

### 10.1 Estado Actual
- Landing page en `golt.dev` con ejemplos básicos.
- Sin API reference completa.
- Sin ejemplos de proyectos reales.

### 10.2 Objetivo
- DX comparable a Deno o Bun: zero-config, errores claros, autocompletado perfecto.

### 10.3 Plan por Fases

#### Fase 10A: Tipados TypeScript (1 semana)
- [ ] Publicar `@types/golt` o tipados en npm.
- [ ] Asegurar que VS Code extension los cargue correctamente.
- [ ] Tipados para todas las APIs nativas.

#### Fase 10B: Mensajes de Error Mejorados (2 semanas)
- [ ] Stack traces que muestren líneas de TypeScript original (source maps).
- [ ] Errores de runtime con sugerencias de corrección.
- [ ] Validación de tipos en runtime para `ctx.ValidateBody` (mejorar API actual).

#### Fase 10C: Plantillas y Ejemplos (1 semana)
- [ ] `golt init --template=api-rest`
- [ ] `golt init --template=graphql`
- [ ] `golt init --template=cli-tool`
- [ ] Repositorio de ejemplos: auth, CRUD, WebSocket, etc.

### 10.4 Dependencias
- **Independiente**.

---

## Roadmap Consolidado: Orden de Ejecución Recomendado

```
Mes 1:
  ├── Módulo 8: Abrir código (inmediato)
  ├── Módulo 1: Spike v8go (semanas 1–2)
  └── Módulo 5: Flags de permisos (semanas 3–4)

Mes 2:
  ├── Módulo 1: Abstracción del motor + integración APIs
  └── Módulo 3: Web Crypto + Web Streams

Mes 3:
  ├── Módulo 2: Cache npm + bundler básico
  └── Módulo 4: PostgreSQL + MySQL

Mes 4:
  ├── Módulo 2: Polyfills Node.js
  ├── Módulo 6: Test runner
  └── Módulo 7: Benchmark suite

Mes 5:
  ├── Módulo 6: Debugger + Linter + Formatter
  ├── Módulo 3: WebSocket + EventTarget
  └── Módulo 9: Multiplataforma + Lambda

Mes 6:
  ├── Módulo 10: DX final, templates, docs
  └── Módulo 7: Optimizaciones + publicación de benchmarks
```

---

## Métricas de Éxito por Módulo

| Módulo | Métrica de éxito | Target |
|--------|-----------------|--------|
| 1 (Motor) | HTTP req/s vs Node.js | ≥ 70% del throughput de Node.js |
| 2 (npm) | % top 100 librerías funcionando | ≥ 80% |
| 3 (Web Standards) | APIs implementadas | fetch, Web Crypto, Streams, WebSocket |
| 4 (DB) | Drivers soportados | SQLite, PostgreSQL, MySQL, Redis |
| 5 (Seguridad) | Flags de permiso | deny-by-default en filesystem, red, env |
| 6 (Toolchain) | Comandos disponibles | test, debug, fmt, lint, check |
| 7 (Perf) | Benchmarks públicos | CI corriendo en cada PR |
| 8 (Open Source) | Métricas de comunidad | ≥ 100 stars, ≥ 5 contribuidores |
| 9 (Deploy) | Plataformas soportadas | Linux, macOS, Windows, Docker, Lambda |
| 10 (DX) | Templates disponibles | ≥ 5 templates oficiales |

---

## Notas Técnicas para el Agente

### Sobre v8go vs goja
- **v8go** requiere CGO. Esto rompe la pureza de Go pero es necesario para rendimiento profesional.
- Considerar **v8go-lite** o forks si el repo original está desactualizado (último commit en v8go fue hace un tiempo; verificar estado en 2026).
- Alternativa a v8go: **QuickJS** via `github.com/second-state/quickjs-go` (más ligero, sin JIT pero más rápido que goja, puro Go bindings).
- Alternativa extrema: usar **Deno** como base y reescribir Golt como un framework sobre Deno (pero esto pierde la identidad de "Go-hosted").

### Sobre el Event Loop
- Con goja, el event loop es trivial (Go goroutines + channels).
- Con v8go, hay que sincronizar el event loop de V8 con el de Go. V8 no es thread-safe por isolate; usar un solo hilo para JS y delegar I/O a goroutines.

### Sobre Bundling
- esbuild es excelente y rápido. Mantenerlo.
- El desafío es la resolución de módulos npm, no el bundling en sí.

---

## Conclusión Ejecutiva para el Agente

Golt tiene una filosofía sólida (API mínima, TypeScript sobre Go), pero le faltan **tres pilares críticos** para ser profesional:

1. **Motor con JIT (V8 via v8go)** — sin esto, el rendimiento es un joke.
2. **Ecosistema npm** — sin esto, nadie puede usar librerías reales.
3. **Código abierto y comunidad** — sin esto, no hay confianza ni contribuciones.

Los demás módulos (DB, seguridad, toolchain, etc.) son importantes pero **secundarios**. Si se resuelven los 3 críticos, Golt pasa de "toy" a "viable alternative".

**Recomendación de orden:** Módulo 8 → Módulo 1 → Módulo 2 → el resto en paralelo.
