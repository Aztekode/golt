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
			engine.Register(runtime.InitHttp)

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
*/
declare namespace Golt {
	export const env: Record<string, string | undefined>;
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
