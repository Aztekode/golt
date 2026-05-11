package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/atrox39/golt/runtime"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

const (
	Reset   = "\033[0m"
	Cyan    = "\033[36;1m"
	Green   = "\033[32;1m"
	Yellow  = "\033[33;1m"
	Magenta = "\033[35;1m"
	Red     = "\033[31;1m"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:     "golt",
		Short:   "Golt Runtime - TS/JS Backend Engine",
		Long:    fmt.Sprintf("%s[Golt] Runtime%s\nA blazing fast backend engine to run TypeScript/JavaScript directly on Go.", Cyan, Reset),
		Version: "1.0.0",
	}

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Golt project",
		Run: func(cmd *cobra.Command, args []string) {
			initProject()
		},
	}

	var runCmd = &cobra.Command{
		Use:   "run [filename.ts]",
		Short: "Run a file in the Golt environment",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filename := args[0]
			fmt.Printf("%s[Golt] * Starting %s...%s\n", Cyan, filename, Reset)

			engine := runtime.NewEngine()
			engine.Register(runtime.InitConsole)
			engine.Register(runtime.InitEnv)
			engine.Register(runtime.InitLogger)
			engine.Register(runtime.InitHttp)
			engine.Register(runtime.InitDB)
			engine.Register(runtime.InitFs)
			engine.Register(runtime.InitFetch)
			engine.Register(runtime.InitCrypto)

			if err := engine.RunFile(filename); err != nil {
				fmt.Printf("%s[Golt] [ERROR] Execution error: %v%s\n", Red, err, Reset)
				os.Exit(1)
			}
		},
	}

	var watchCmd = &cobra.Command{
		Use:   "watch [filename.ts]",
		Short: "Run a file and automatically restart on changes",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filename := args[0]
			watchApp(filename)
		},
	}

	rootCmd.AddCommand(initCmd, runCmd, watchCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("%s[Golt] [ERROR] Unrecognized command.%s\n", Red, Reset)
		os.Exit(1)
	}
}

func initProject() {
	dtsContents := `/***
 * Golt Runtime - TS/JS Backend Engine Global Definitions
 ***/
declare interface FetchHeaders {
  get(name: string): string | null;
}

declare interface FetchResponse {
  ok: boolean;
  status: number;
  statusText: string;
  headers: FetchHeaders;
  text(): Promise<string>;
  json<T = any>(): Promise<T>;
}

declare interface FetchOptions {
  method?: "GET" | "POST" | "PUT" | "DELETE" | "PATCH";
  headers?: Record<string, string>;
  body?: string;
  timeout?: number;
}

declare function fetch(url: string, options?: FetchOptions): Promise<FetchResponse>;

declare namespace Golt {
  export const env: Record<string, string | undefined>;

  export type SchemaType = "string" | "number" | "boolean";

  export type InferType<T> = T extends "string"
    ? string
    : T extends "number"
    ? number
    : T extends "boolean"
    ? boolean
    : never;

  export type InferSchema<T extends Record<string, SchemaType>> = {
    [K in keyof T]: InferType<T[K]>;
  };

  export type Next = () => void;
  export type Middleware = (c: Context, next: Next) => void;

  export interface LoggerConfig {
    format?: "dev" | "tiny" | "json";
  }

  export type DbDialect = "sqlite" | "postgres" | "mysql" | "sqlserver";

  export interface DatabaseClient {
    query<T = any>(sql: string, ...args: any[]): Promise<T[]>;
    close(): void;
  }

  export interface Database {
    connect(dialect: DbDialect, connectionString: string): DatabaseClient; 
    query<T = any>(sql: string, ...args: any[]): Promise<T[]>; 
  }

  export interface Context {
    Method(): string;
    Url(): string;
    Param(name: string): string;
    GetHeader(key: string): string;
    SetHeader(key: string, value: string): void;
    Set(key: string, value: any): void;
    Get<T = any>(key: string): T | undefined;
    Query(key: string): string;
    Status(code: number): Context;
    Send(body: string): void;
    Json(data: any): void;
    ValidateBody<T extends Record<string, SchemaType>>(
      schema: T,
    ): InferSchema<T> | null;
  }

  export interface Fs {
    readFile(path: string): string;
    writeFile(path: string, content: string): void;
  }

  export interface Crypto {
    hash(password: string, cost?: number): Promise<string>;
    compare(password: string, hash: string): Promise<boolean>;
  }

  export interface Jwt {
    sign(payload: Record<string, any>, secret: string, expHours?: number): string;
    verify<T = Record<string, any>>(token: string, secret: string): T | null;
  }

  export interface AppInstance {
    use(middleware: Middleware): AppInstance;
    get(path: string, handler: (c: Context) => void): AppInstance;
    post(path: string, handler: (c: Context) => void): AppInstance;
    put(path: string, handler: (c: Context) => void): AppInstance;
    delete(path: string, handler: (c: Context) => void): AppInstance;
    static(prefix: string, dirPath: string, spa?: boolean): AppInstance;
    notFound(handler: (c: Context) => void): AppInstance;
    serve(port: number): void;
  }

  export function App(): AppInstance;
  export const db: Database;
  export const fs: Fs;
  export const crypto: Crypto;
  export const jwt: Jwt;

  export function logger(config?: LoggerConfig): Middleware;
}`

	tsConfigContent := `{
	"compilerOptions": {
		"target": "ESNext",
		"module": "ESNext",
		"moduleResolution": "node",
		"strict": true,
		"esModuleInterop": true,
		"skipLibCheck": true,
		"forceConsistentCasingInFileNames": true
	},
	"include": [
		"**/*.ts",
		"golt.d.ts"
	]
}
`

	appTsContent := `console.log('Hello, Golt!');`

	os.WriteFile("golt.d.ts", []byte(dtsContents), 0644)
	os.WriteFile("tsconfig.json", []byte(tsConfigContent), 0644)
	os.WriteFile("app.ts", []byte(appTsContent), 0644)

	fmt.Printf("%s[Golt] * Proyecto inicializado con éxito. Archivos creados:%s\n", Green, Reset)
	fmt.Printf("  %s+%s golt.d.ts (Definiciones Globales)\n", Green, Reset)
	fmt.Printf("  %s+%s tsconfig.json (Configuración de TypeScript)\n", Green, Reset)
	fmt.Printf("  %s+%s app.ts (Punto de entrada)\n\n", Green, Reset)

	fmt.Printf("%s[Golt] Ejecuta tu proyecto con:%s\n", Cyan, Reset)
	fmt.Printf("  golt run app.ts\n")
	fmt.Printf("  golt watch app.ts\n")
}

func watchApp(filePath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("%s[Golt] [ERROR] Error starting watcher: %v%s\n", Red, err, Reset)
		return
	}
	defer watcher.Close()

	if err := watcher.Add("."); err != nil {
		fmt.Printf("%s[Golt] [ERROR] Error watching directory: %v%s\n", Red, err, Reset)
		return
	}

	fmt.Printf("%s[Golt] [WATCH] Watch mode activated. Waiting for changes in %s...%s\n", Magenta, filePath, Reset)

	var cmd *exec.Cmd

	startProcess := func() {
		if cmd != nil && cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
		}

		cmd = exec.Command(os.Args[0], "run", filePath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			fmt.Printf("%s[Golt] [ERROR] Error starting process: %v%s\n", Red, err, Reset)
		}
	}

	startProcess()

	var lastEvent time.Time
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if strings.HasSuffix(event.Name, ".js") || strings.HasSuffix(event.Name, ".ts") {
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 && time.Since(lastEvent) > 500*time.Millisecond {
					lastEvent = time.Now()
					fmt.Print("\033[2J\033[H")
					fmt.Printf("%s[Golt] [RELOAD] Changes detected. Restarting server...%s\n", Yellow, Reset)
					startProcess()
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("%s[Golt] [ERROR] Watcher error: %v%s\n", Red, err, Reset)
		}
	}
}
