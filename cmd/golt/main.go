package main

import (
	"fmt"
	"os"

	"github.com/atrox39/golt/runtime"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "golt",
		Short: "Golt Runtime - TS/JS Backend Engine",
		Long:  `Golt Runtime is a TypeScript/JavaScript backend engine that allows you to run TypeScript code in as Node.js environment.`,
	}

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize new Golt project",
		Run: func(cmd *cobra.Command, args []string) {
			initProject()
		},
	}

	var runCmd = &cobra.Command{
		Use:   "run [filename.ts]",
		Short: "Run file in Golt environment",
		Run: func(cmd *cobra.Command, args []string) {
			filename := args[0]
			engine := runtime.NewEngine()
			engine.Register(runtime.InitConsole)
			engine.Register(runtime.InitEnv)
			engine.Register(runtime.InitLogger)
			engine.Register(runtime.InitHttp)
			engine.Register(runtime.InitDB)
			engine.Register(runtime.InitFs)
			engine.Register(runtime.InitFetch)

			engine.RunFile(filename)
		},
	}

	rootCmd.AddCommand(initCmd, runCmd)

	if err := rootCmd.Execute(); err != nil {
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

  export interface Database {
    connect(dialect: DbDialect, connectionString: string): void;
    query<T = any>(sql: string, ...args: any[]): Promise<T[]>;
  }

  export interface Context {
    Method(): string;
    Url(): string;
    Param(name: string): string;
    GetHeader(key: string): string;
    SetHeader(key: string, value: string): void;
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

  export interface AppInstance {
    use(middleware: Middleware): AppInstance; // <-- Añadido
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

  export function logger(config?: LoggerConfig): Middleware; // <-- Añadido
}
`
	os.WriteFile("golt.d.ts", []byte(dtsContents), 0644)

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
	os.WriteFile("tsconfig.json", []byte(tsConfigContent), 0644)

	appTsContent := `console.log('Hello, Golt!');`

	os.WriteFile("app.ts", []byte(appTsContent), 0644)

	fmt.Println("Golt Project Initialized. Files created:")
	fmt.Println("- golt.d.ts (Global Definitions)")
	fmt.Println("- tsconfig.json (TypeScript Configuration)")
	fmt.Println("- app.ts (Entry Point)")
	fmt.Println("Run the project with:")
	fmt.Println("golt run app.ts")
}
