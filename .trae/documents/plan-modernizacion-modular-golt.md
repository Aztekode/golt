# Plan de Ejecucion Modular de Golt

## Summary
- Objetivo: convertir `PLAN.md` en un plan de ejecucion realista, modular y aterrizado al estado actual del repo.
- Orden acordado: comenzar por **Modulo 8 (Gobernanza/Open Source)** y mantener un **roadmap completo** que prepare la transicion a Modulo 1, Modulo 2 y el resto.
- Enfoque: ejecutar por modulos, con entregables pequeños y verificables, evitando cambios masivos sin puntos de control.

## Current State Analysis

### Estado real del repo
- Runtime actual basado en `goja` + `goja_nodejs` en [engine.go](file:///c:/Users/cortega/Development/personal/golt-project/golt/runtime/engine.go).
- API nativa actual expuesta desde `runtime/` con modulos como:
  - HTTP en [http.go](file:///c:/Users/cortega/Development/personal/golt-project/golt/runtime/http.go)
  - DB en [db.go](file:///c:/Users/cortega/Development/personal/golt-project/golt/runtime/db.go)
  - FS en [fs.go](file:///c:/Users/cortega/Development/personal/golt-project/golt/runtime/fs.go)
  - Crypto/JWT en [crypto.go](file:///c:/Users/cortega/Development/personal/golt-project/golt/runtime/crypto.go)
- CLI actual concentrado en [main.go](file:///c:/Users/cortega/Development/personal/golt-project/golt/cmd/golt/main.go) con comandos `init`, `run`, `watch`, `verify-release`.
- Sistema de release ya existente:
  - workflow en [release.yml](file:///c:/Users/cortega/Development/personal/golt-project/golt/.github/workflows/release.yml)
  - instalador en [installer.iss](file:///c:/Users/cortega/Development/personal/golt-project/golt/installer.iss)
  - verificacion de release en `internal/verify/`
- Documentacion actual existente:
  - [README.md](file:///c:/Users/cortega/Development/personal/golt-project/golt/README.md)
  - [CHANGELOG.md](file:///c:/Users/cortega/Development/personal/golt-project/golt/CHANGELOG.md)
  - sitio docs en [index.html](file:///c:/Users/cortega/Development/personal/golt-project/golt/docs/index.html)
  - ejemplos en `examples/`

### Estado del Modulo 8 hoy
- Parcialmente avanzado:
  - Existe repo publico
  - Existe `README.md`
  - Existe `CHANGELOG.md`
  - Existe workflow de releases
- Faltantes claros:
  - No existe `LICENSE`
  - No existe `CONTRIBUTING.md`
  - No existe `CODE_OF_CONDUCT.md`
  - No existe `SECURITY.md`
  - No existe roadmap publico corto/largo separado de `PLAN.md`
  - No existe template de issues/PR ni lineamientos de releases/soporte

### Estado del Modulo 1 hoy
- No existe abstraccion `JSEngine`.
- Todo el runtime depende de `goja` directamente:
  - [engine.go](file:///c:/Users/cortega/Development/personal/golt-project/golt/runtime/engine.go)
  - modulos `runtime/*` asumen `*goja.Runtime`
- No existe spike de `v8go`.
- No existe feature flag `--engine=...`.

### Estado del Modulo 2 hoy
- No existe `package.json`, `golt.lock`, `golt install`, cache npm ni resolucion de paquetes externos.
- `esbuild` solo bundlea el entrypoint actual.

### Restricciones reales detectadas
- Preferencia registrada del usuario: evitar “parchar dependencias externas” dentro del monorepo si se puede evitar.
- El repo ya esta liberando `v1.0.3`; por lo tanto, las mejoras grandes deben preservar una ruta de estabilizacion y no romper release sin una capa de compatibilidad.
- La migracion a `v8go` implica CGO y complejidad real de cross-compilation; no debe asumirse como cambio inmediato sin spike.

## Proposed Changes

### Fase 0: Base de ejecucion y governance minima (arranque del Modulo 8)

**Objetivo**
- Dejar el proyecto con un esqueleto open source profesional antes de cambios profundos de arquitectura.

**Archivos a crear/actualizar**
- Nuevo: `LICENSE`
- Nuevo: `CONTRIBUTING.md`
- Nuevo: `CODE_OF_CONDUCT.md`
- Nuevo: `SECURITY.md`
- Nuevo: `.github/ISSUE_TEMPLATE/*`
- Nuevo: `.github/pull_request_template.md`
- Nuevo: `ROADMAP.md`
- Actualizar: [README.md](file:///c:/Users/cortega/Development/personal/golt-project/golt/README.md)
- Actualizar: [CHANGELOG.md](file:///c:/Users/cortega/Development/personal/golt-project/golt/CHANGELOG.md)

**Que hacer**
1. Elegir licencia permisiva (MIT recomendada por simplicidad).
2. Definir politicas de contribucion y vulnerabilidades.
3. Separar:
   - `PLAN.md` como plan interno/arquitectonico largo
   - `ROADMAP.md` como roadmap publico resumido
4. Actualizar `README.md` con:
   - estado del proyecto
   - soporte actual real
   - roadmap corto
   - como contribuir
5. Crear templates de issues/PR para canalizar trabajo por modulo.

**Resultado esperado**
- Modulo 8 queda suficientemente cubierto para que el proyecto sea presentable y mantenible.

### Fase 1: Preparacion arquitectonica para Modulo 1 (sin migrar aun el motor)

**Objetivo**
- Reducir el acoplamiento directo con `goja` antes del spike con `v8go`.

**Archivos principales**
- Nuevo: `engine/` o `internal/engine/` para la capa de abstraccion
- Refactorizar: [engine.go](file:///c:/Users/cortega/Development/personal/golt-project/golt/runtime/engine.go)
- Ajustar: modulos en `runtime/*`
- Ajustar: [main.go](file:///c:/Users/cortega/Development/personal/golt-project/golt/cmd/golt/main.go)

**Que hacer**
1. Definir interfaz base de engine:
   - evaluacion
   - globals
   - lifecycle
   - colas/event loop
2. Crear wrapper `GojaEngine` que preserve comportamiento actual.
3. Mover el codigo actual de `runtime/engine.go` a una implementacion concreta en vez de ser la unica ruta.
4. Introducir feature flag en CLI:
   - `--engine=goja`
   - `--engine=v8go` (deshabilitado o experimental mientras no exista implementacion viable)
5. Documentar las diferencias entre:
   - engine loop
   - promesas
   - exposicion de APIs nativas

**Resultado esperado**
- El proyecto sigue funcionando con `goja`, pero ya no depende de el en cada punto de entrada.

### Fase 2: Spike real de viabilidad de Modulo 1 (v8go)

**Objetivo**
- Validar si `v8go` es viable en 2026 para Windows/Linux/macOS en este repo.

**Archivos principales**
- Nuevo: `cmd/spikes/v8go-smoke/` o `spikes/v8go/`
- Nuevo: `docs/adr/ADR-001-engine-selection.md`
- Ajustar temporalmente: `go.mod`

**Que hacer**
1. Resolver el estado actual de `v8go` en 2026:
   - mantenimiento del repo
   - compatibilidad con Go actual
   - soporte de plataformas requeridas
2. Crear smoke test minimo:
   - eval JS
   - set/get globals
   - promesas/microtasks
3. Medir:
   - cold start
   - complejidad de compilacion
   - tamaño de binario
4. Escribir un ADR con decision:
   - seguir con `v8go`
   - considerar `quickjs-go`
   - mantener `goja` temporalmente

**Resultado esperado**
- Decision tecnica basada en evidencia antes de reescribir APIs nativas.

### Fase 3: Cierre del Modulo 1 (solo si el spike da go)

**Objetivo**
- Habilitar una ruta experimental de ejecucion sobre un segundo motor.

**Archivos principales**
- Implementacion `V8Engine`
- Adaptaciones puntuales en `runtime/*`
- Tests de compatibilidad por API

**Que hacer**
1. Implementar engine alterno.
2. Exponer globals equivalentes (`Golt.App`, `Golt.db`, `Golt.fs`, etc.).
3. Confirmar que `async/await` y promesas funcionen con el modelo hibrido.
4. Mantener `goja` como default hasta tener benchmarks.

**Resultado esperado**
- `golt run --engine=v8go` experimental, sin romper `goja`.

### Fase 4: Modulo 2 (npm y sistema de modulos)

**Objetivo**
- Dejar de depender de un unico entrypoint bundlado y habilitar dependencias reales.

**Archivos principales**
- Nuevo: `internal/npm/*` o `pkg/npm/*`
- Nuevo: `internal/lockfile/*`
- Ajustar: [main.go](file:///c:/Users/cortega/Development/personal/golt-project/golt/cmd/golt/main.go)
- Ajustar: [engine.go](file:///c:/Users/cortega/Development/personal/golt-project/golt/runtime/engine.go) o capa nueva de bundling
- Nuevo: docs para `golt install`

**Que hacer**
1. Introducir `golt install`.
2. Crear cache local de paquetes en home del usuario.
3. Resolver semver basica y lockfile.
4. Hacer que esbuild resuelva desde cache/local node_modules.
5. Postergar polyfills complejos de Node hasta tener:
   - decision estable del motor
   - capa de APIs mas limpia

**Resultado esperado**
- Primer soporte reproducible para dependencias externas, aunque parcial.

### Fase 5: Modulos paralelizables despues del bloque critico

**Objetivo**
- Ejecutar trabajo paralelo una vez que el proyecto tenga base publica, engine strategy y sistema de modulos inicial.

**Bloques recomendados**
1. **Modulo 5 (Seguridad)**
   - puede empezar antes de completar V8
   - flags de permisos y enforcement en `fs`, `env`, `fetch`
2. **Modulo 10 (DX)**
   - errores claros, sourcemaps, templates, tipados
   - ya hay avance parcial en v1.0.3
3. **Modulo 4 (DB)**
   - consolidar postgres/mysql/redis y transacciones
4. **Modulo 6 (Toolchain)**
   - `golt test`, `golt check`, `golt fmt`, `golt lint`
5. **Modulo 7 (Benchmarking)**
   - despues del spike/decision de engine
6. **Modulo 9 (Deployment)**
   - despues de estabilizar release y compatibilidad multiplataforma

## Assumptions & Decisions
- Decision tomada con el usuario:
  - iniciar por **Modulo 8**
  - preparar un **roadmap completo**
- El plan NO asume que `v8go` sea automaticamente la respuesta correcta; primero debe pasar por spike y ADR.
- Se preserva la version de producto actual y la rama de releases existente mientras se hacen cambios modulares.
- Los cambios deben privilegiar:
  - compatibilidad incremental
  - feature flags
  - evidencia tecnica antes de migraciones profundas

## Verification

### Verificacion del Modulo 8
- Existen `LICENSE`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`, `ROADMAP.md`.
- `README.md` referencia claramente roadmap, soporte y contribucion.
- Existen templates de issue/PR.

### Verificacion del Modulo 1
- Existe una interfaz de engine aislada del runtime actual.
- `goja` corre a traves del wrapper nuevo sin regresiones.
- Existe spike reproducible y ADR de decision para `v8go`/alternativa.

### Verificacion del Modulo 2
- `golt install` resuelve al menos una dependencia simple.
- Existe cache local y lockfile reproducible.
- Esbuild resuelve modulos externos sin hacks manuales.

### Verificacion transversal
- Cada modulo se entrega con:
  - cambios de codigo
  - documentacion
  - pasos de prueba
  - decision de go/no-go si aplica

## Orden Recomendado de Ejecucion
1. Modulo 8 completo (gobernanza/open source)
2. Preparacion arquitectonica de Modulo 1
3. Spike de viabilidad de motor
4. Decision ADR
5. Primer corte de Modulo 2
6. Modulo 5 y Modulo 10 en paralelo
7. Modulo 4
8. Modulo 6
9. Modulo 7
10. Modulo 9

## Primer Entregable Recomendado
- Ejecutar primero un **Modulo 8 completo** con entregables publicos y luego entrar a **Modulo 1 Fase 1 + Spike** en la siguiente iteracion.
