# Plan para v1.0.3 (Golt Runtime)

## Summary
- Objetivo: preparar una v1.0.3 enfocada en (1) detección de versiones (local + remoto), (2) mejorar errores/sourcemaps de TypeScript en runtime, (3) mejorar instalador Inno Setup con branding e instalación opcional de ejemplos, (4) firma/validación para distinguir instaladores oficiales de Aztekode/golt.
- Resultado esperado: versión única y consistente en CLI/Docker/installer, runtime con errores más accionables, instalador más “producto” (iconos, tareas opcionales, limpieza de PATH) y firmado, carpeta `examples/` empaquetable, y mecanismo de verificación (firma Windows + hashes firmados).

## Current State Analysis

### Versionado
- La versión del CLI está hardcodeada en Cobra: [cmd/golt/main.go](file:///c:/Users/cortega/Development/personal/golt-project/golt/cmd/golt/main.go#L25-L32) (`Version: "1.0.2"`).
- El instalador también hardcodea la versión: [installer.iss](file:///c:/Users/cortega/Development/personal/golt-project/golt/installer.iss#L1-L12) (`AppVersion=1.0.2`, `OutputBaseFilename=...1.0.2...`).
- Docker intenta inyectar versión con `-ldflags "-X main.version=${VERSION}"`, pero hoy no existe `var version` en `package main`: [Dockerfile](file:///c:/Users/cortega/Development/personal/golt-project/golt/Dockerfile#L11-L16).
- Docs contienen strings con versión (no automatizado): [docs/index.html](file:///c:/Users/cortega/Development/personal/golt-project/golt/docs/index.html#L677-L746).

### Runtime TypeScript (errores)
- El runtime compila usando esbuild con bundle y ejecuta con goja + eventloop: [runtime/engine.go](file:///c:/Users/cortega/Development/personal/golt-project/golt/runtime/engine.go#L32-L81).
- Errores de compilación: solo se devuelve el primer error (`buildResult.Errors[0].Text`) sin ubicación (file:line:col) ni lista completa.
- Errores de runtime: se reporta `goja.Exception.String()`; no se muestran sourcemaps ni trazas en TS.

### Instalador Inno Setup
- Actualmente solo instala `golt.exe` y puede agregar `{app}` al PATH de usuario (HKCU): [installer.iss](file:///c:/Users/cortega/Development/personal/golt-project/golt/installer.iss#L19-L27).
- No hay limpieza del PATH al desinstalar.
- No hay icono `.ico`, ni imágenes del wizard; en el repo solo existe [docs/images/icon.jpg](file:///c:/Users/cortega/Development/personal/golt-project/golt/docs/images/icon.jpg).
- El usuario solicitó modo “Elegible (dialog)” (per-user / per-machine) y generar assets con ffmpeg a partir de `icon.jpg`.

### Firma / oficialidad
- Hoy no hay firma del binario/instalador (no existe configuración de Inno `SignTool` ni pipeline de release).
- Riesgo: cualquiera puede compilar un instalador desde `installer.iss`; la distinción de “oficial” debe venir por firma (Authenticode) y/o assets publicados con hashes firmados.

### Examples
- No existe `examples/` hoy; los ejemplos están en README y el scaffold de `golt init` está hardcodeado en [cmd/golt/main.go](file:///c:/Users/cortega/Development/personal/golt-project/golt/cmd/golt/main.go#L90-L168).

## Proposed Changes

### 1) Fuente única de versión + detección de “actualización local” + check remoto

**Archivos principales**
- [cmd/golt/main.go](file:///c:/Users/cortega/Development/personal/golt-project/golt/cmd/golt/main.go)
- Nuevo: `cmd/golt/version.go` (o similar) dentro de `package main`
- Nuevo: `internal/updatecheck/*` (cliente update + semver mínima + cache)

**Cambios**
1. **Unificar versión del binario**
   - Introducir `var version = "dev"` en `package main` y asignar `rootCmd.Version = version` en lugar del string hardcodeado.
   - Objetivo: que el `-ldflags -X main.version=...` funcione (y también `golt --version`/`golt version` muestre lo real).
2. **Detección local (primer arranque después de update)**
   - Guardar un “estado” en directorio de configuración del usuario (Windows: `%AppData%\\Golt\\state.json` o equivalente con `os.UserConfigDir()`):
     - `last_seen_version`
     - `last_update_check_at` (para el check remoto)
   - En cada arranque del CLI: si `version != last_seen_version` => imprimir mensaje “Actualizado a X (antes Y)” y persistir el nuevo valor.
3. **Check remoto (no-bloqueante, con cache)**
   - Hacer request a GitHub Releases latest (repo `Aztekode/golt`) y comparar con la versión local.
   - Validar “oficialidad” del release (ver sección 5) antes de mostrar “hay update”.
   - Frecuencia: 1 vez cada 24h (configurable).
   - Controles:
     - variable env para deshabilitar (`GOLT_NO_UPDATE_CHECK=1`)
     - flag `--no-update-check` (persistente en rootCmd)
     - timeouts cortos (ej. 2–3s) y sin romper el comando si falla.
   - Output: aviso de “Nueva versión disponible” con URL.

**Decisiones ya tomadas**
- El usuario pidió “Ambas”: detección local + check remoto.

### 2) Mejoras TypeScript en runtime (errores y sourcemaps)

**Archivos principales**
- [runtime/engine.go](file:///c:/Users/cortega/Development/personal/golt-project/golt/runtime/engine.go)
- (Opcional) nuevo helper: `runtime/diagnostics.go`

**Cambios**
1. **Errores de compilación más completos**
   - Iterar todos los `buildResult.Errors` y construir un error multi-línea con:
     - archivo + línea + columna (si existe `Location`)
     - `Text` + `LineText`
     - sugerencia básica cuando aplique (si `Notes` existe)
2. **Generar sourcemaps**
   - Habilitar sourcemap inline en esbuild (`Sourcemap: Inline`) y `SourcesContent` incluido.
   - Mantener `Write: false`, pero conservar el JS con inline map para ejecución.
3. **Errores de runtime más accionables**
   - Cuando ocurra `goja.Exception`, incluir stack trace si está disponible (en vez de solo `String()`).
   - (Si goja no consume sourcemaps automáticamente) Paso 2: implementar un mapeo mínimo:
     - extraer el sourcemap inline de `compiledCode`
     - mapear 1–N frames de stack (best-effort) usando `github.com/go-sourcemap/sourcemap` (ya está en go.mod como indirect)
   - Mantenerlo “best effort”: si falla el parsing/mapeo, mostrar stack original.

**Criterio de éxito**
- Un error TS debe mostrar file:line:col y el `LineText`.
- Un panic/runtime error debe mostrar stack con mayor contexto (idealmente con referencias al archivo TS original cuando sea posible).

### 3) Instalador Inno Setup: branding, tasks y limpieza

**Archivos principales**
- [installer.iss](file:///c:/Users/cortega/Development/personal/golt-project/golt/installer.iss)
- Nuevo: `assets/installer/*` (iconos e imágenes generadas desde `docs/images/icon.jpg`)

**Cambios**
1. **Hacer instalación “Elegible (dialog)”**
   - Configurar `PrivilegesRequiredOverridesAllowed=dialog`.
   - Definir rutas:
     - per-user: `{localappdata}\\Programs\\Golt` (y PATH en HKCU)
     - per-machine: `{pf}\\Golt` (y PATH en HKLM)
2. **Branding del instalador**
   - Generar desde `docs/images/icon.jpg`:
     - `assets/installer/golt.ico` (para `SetupIconFile` y shortcuts)
     - `assets/installer/wizard.bmp` y `wizard-small.bmp` (para `WizardImageFile`/`WizardSmallImageFile`)
   - Generación pedida por el usuario: usar `ffmpeg` durante el proceso de release (y versionar los resultados en el repo).
3. **Tasks nuevas**
   - `envPath`: mantener (PATH).
   - `examples`: instalar ejemplos a `{app}\\examples` (opcional).
   - (Opcional) `desktopicon`: shortcut escritorio.
4. **[Files] y [Icons]**
   - Agregar:
     - `[Files]` para `examples\\*` con `recursesubdirs` y `Tasks: examples`
     - `[Icons]` para Start Menu y/o Desktop (condicionado a task)
5. **PATH: alta y baja limpia**
   - Mejorar escritura del PATH para manejar caso de PATH vacío (evitar `;{app}` doble / `;;`).
   - Implementar remoción en uninstall (en `[Code]` usando `CurUninstallStepChanged` o equivalente), quitando exactamente `{app}` del PATH (HKCU y HKLM según modo).
6. **Metadatos**
   - Definir `AppId` estable (GUID).
   - Definir `UninstallDisplayIcon` apuntando al `.ico` o al exe.

### 4) Carpeta `examples/` (contenido y empaquetado)

**Archivos principales**
- Nuevo: `examples/` con 2–4 ejemplos mínimos, por ejemplo:
  - `examples/hello-http/app.ts` (hello world + logger)
  - `examples/sqlite/app.ts` (connect + query + exec)
  - `examples/static-spa/app.ts` (static + spaFallback)
- (Opcional) `examples/README.md` explicando cómo correrlos.

**Criterio de éxito**
- El instalador muestra una opción para instalar ejemplos y, si se marca, copia la carpeta a `{app}\\examples`.

### 5) Instalador oficial firmado + verificación (firma Windows + hashes firmados)

**Objetivo**
- Que el instalador oficial publicado por Aztekode/golt:
  - esté firmado con Authenticode (PFX) y timestamp.
  - publique un manifiesto de hashes (SHA-256) firmado (para verificación adicional e integridad).
- Que el CLI pueda validar “oficialidad” de un release (al menos a nivel de manifiesto firmado) cuando hace el check remoto.

**Archivos principales**
- [installer.iss](file:///c:/Users/cortega/Development/personal/golt-project/golt/installer.iss)
- Nuevo: `.github/workflows/release.yml` (build + firma + release assets)
- Nuevo: `internal/verify/*` (verificación de manifiesto firmado)
- Nuevo: `internal/release/SHA256SUMS` (formato generado en CI, no mantenido a mano)
- Nuevo: `docs/` o `README.md` con instrucciones para verificar firma Windows + hashes

**Cambios**
1. **Firma Authenticode en Inno Setup**
   - Configurar firma del instalador (y opcionalmente del uninstaller) con `SignTool`.
   - Parametrizar por variables de entorno/CI:
     - ruta a `signtool.exe`
     - PFX (inyectado como secreto; nunca en repo) + password
     - URL de timestamp (ej. RFC3161)
   - El build local sin secretos debe poder compilar sin firma (fallback); el build “oficial” en CI debe fallar si no se firma.
2. **Firma del binario `golt.exe`**
   - Firmar `golt.exe` antes de empaquetarlo en el setup para que:
     - el instalador muestre publisher “Aztekode”
     - el binario instalado también quede firmado
3. **Manifiesto de hashes firmado**
   - En CI, generar `SHA256SUMS` de todos los artefactos publicados (setup + binarios por plataforma).
   - Firmar `SHA256SUMS` con una llave Ed25519 (minisign compatible):
     - `SHA256SUMS.minisig`
     - llave privada en secreto CI; llave pública versionada en repo (para verificación).
4. **Verificación en CLI (oficialidad)**
   - Agregar comando `golt verify-release` (o `golt verify <path>`):
     - valida SHA-256 contra `SHA256SUMS`
     - valida firma `SHA256SUMS.minisig` contra la llave pública embebida
   - El check remoto usará esta verificación para “confiar” en la versión/remotos:
     - descarga `SHA256SUMS` + `.minisig` del release
     - valida la firma antes de sugerir actualizar
   - Nota: la verificación de Authenticode desde Go sería específica de Windows y puede quedar como “verificación manual” documentada (Properties → Digital Signatures).

**Criterio de éxito**
- El instalador oficial muestra firma válida (publisher Aztekode) en Windows.
- Existe `SHA256SUMS` + `SHA256SUMS.minisig` en el release.
- `golt verify-release` confirma la validez de un asset descargado cuando el manifiesto y firma corresponden.

## Assumptions & Decisions
- La v1.0.3 seguirá siendo runtime “no-Node”: esbuild bundle + goja + eventloop (sin `node_modules` runtime).
- El check remoto será aviso (no auto-update) para mantenerlo simple y seguro; un `golt upgrade` quedaría fuera de alcance.
- Los assets del instalador se generan desde `docs/images/icon.jpg` usando ffmpeg y se guardan en el repo para builds reproducibles.
- Firma oficial: el usuario eligió Authenticode con certificado PFX, y “Combinado” como validación (firma Windows + hashes firmados).

## Verification
- CLI:
  - `golt --version` muestra la versión inyectada por build.
  - Primer arranque tras cambiar versión muestra mensaje de “actualizado”.
  - Check remoto: si no hay red, no debe fallar el comando (solo silencioso/aviso).
- Runtime:
  - Ejecutar un `.ts` con error de types debe mostrar file:line:col + line text.
  - Provocar error en runtime (throw) y verificar stack trace más informativo.
- Instalador:
  - Instalar per-user y per-machine desde el mismo instalador (dialog).
  - Si se marca PATH, se agrega y se remueve al desinstalar.
  - Si se marca examples, se copian a `{app}\\examples`.
- Firma/validación:
  - `golt.exe` y el instalador están firmados en el build oficial.
  - El release incluye `SHA256SUMS` + `SHA256SUMS.minisig`.
  - `golt verify-release` valida correctamente un asset oficial.
