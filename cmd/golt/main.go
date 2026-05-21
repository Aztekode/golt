package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Aztekode/golt/runtime"
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
		Version: "1.0.2",
	}

	var initCmd = &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize a new Golt project",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectName := args[0]

			if err := initProject(projectName); err != nil {
				fmt.Printf("%s[Golt] [ERROR] %v%s\n", Red, err, Reset)
				os.Exit(1)
			}
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

func initProject(projectName string) error {
	projectName = strings.TrimSpace(projectName)

	if projectName == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	if strings.Contains(projectName, string(os.PathSeparator)) {
		return fmt.Errorf("project name must be a folder name, not a path")
	}

	projectPath := filepath.Clean(projectName)

	if _, err := os.Stat(projectPath); err == nil {
		return fmt.Errorf("folder %q already exists", projectPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("could not check project folder: %w", err)
	}

	vscodePath := filepath.Join(projectPath, ".vscode")
	extensionsPath := filepath.Join(vscodePath, "extensions.json")

	if err := os.MkdirAll(vscodePath, 0755); err != nil {
		return fmt.Errorf("could not create project folders: %w", err)
	}

	appTsContent := `const app = Golt.App();

app.use(Golt.logger({ format: "dev" }));

app.get("/", (ctx) => {
  ctx.Json({
    message: "Hello from Golt!",
    runtime: "golt",
  });
});

app.serve(3000);
`

	goltJsonContent := fmt.Sprintf(`{
  "name": "%s",
  "description": "A Golt Runtime project",
  "version": "0.1.0"
}
`, projectName)

	extensionsJsonContent := `{
  "recommendations": [
    "Aztekode.golt-vscode"
  ]
}
`

	files := map[string]string{
		filepath.Join(projectPath, "app.ts"):    appTsContent,
		filepath.Join(projectPath, "golt.json"): goltJsonContent,
		extensionsPath:                          extensionsJsonContent,
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("could not write %s: %w", path, err)
		}
	}

	fmt.Printf("%s[Golt] * Project initialized successfully:%s\n", Green, Reset)
	fmt.Printf("  %s+%s %s/\n", Green, Reset, projectPath)
	fmt.Printf("  %s+%s app.ts\n", Green, Reset)
	fmt.Printf("  %s+%s golt.json\n", Green, Reset)
	fmt.Printf("  %s+%s .vscode/extensions.json\n\n", Green, Reset)

	fmt.Printf("%s[Golt] Next steps:%s\n", Cyan, Reset)
	fmt.Printf("  cd %s\n", projectPath)
	fmt.Printf("  code .\n")
	fmt.Printf("  golt run app.ts\n")

	return nil
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
