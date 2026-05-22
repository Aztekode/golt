package runtime

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/evanw/esbuild/pkg/api"
)

type NativeModule func(vm *goja.Runtime, e *GoltEngine)

type GoltEngine struct {
	loop    *eventloop.EventLoop
	modules []NativeModule
	Wg      sync.WaitGroup
}

func NewEngine() *GoltEngine {
	engine := &GoltEngine{
		loop: eventloop.NewEventLoop(),
	}

	return engine
}

func (e *GoltEngine) Register(module NativeModule) {
	e.modules = append(e.modules, module)
}

func (e *GoltEngine) RunFile(filename string) error {
	absFilename, err := filepath.Abs(filename)
	if err != nil {
		absFilename = filename
	}

	buildResult := api.Build(api.BuildOptions{
		EntryPoints:    []string{absFilename},
		Bundle:         true,
		Write:          false,
		Format:         api.FormatIIFE,
		Target:         api.ES2015,
		Platform:       api.PlatformNode,
		LogLevel:       api.LogLevelSilent,
		Sourcemap:      api.SourceMapInline,
		SourceRoot:     "",
		AbsWorkingDir:  filepath.Dir(absFilename),
		SourcesContent: api.SourcesContentInclude,
	})

	if len(buildResult.Errors) > 0 {
		var b strings.Builder
		b.WriteString("Compilation error(s):\n")
		for _, e := range buildResult.Errors {
			if e.Location != nil {
				fmt.Fprintf(&b, "- %s:%d:%d: %s\n", e.Location.File, e.Location.Line, e.Location.Column, e.Text)
				if e.Location.LineText != "" {
					fmt.Fprintf(&b, "  %s\n", e.Location.LineText)
				}
			} else {
				fmt.Fprintf(&b, "- %s\n", e.Text)
			}
			for _, n := range e.Notes {
				if n.Location != nil {
					fmt.Fprintf(&b, "  note: %s:%d:%d: %s\n", n.Location.File, n.Location.Line, n.Location.Column, n.Text)
				} else {
					fmt.Fprintf(&b, "  note: %s\n", n.Text)
				}
			}
		}
		return fmt.Errorf("%s", strings.TrimRight(b.String(), "\n"))
	}

	compiledCode := string(buildResult.OutputFiles[0].Contents)

	e.loop.Start()
	defer e.loop.Stop()

	var scriptWg sync.WaitGroup
	scriptWg.Add(1)

	var runErr error

	e.loop.RunOnLoop(func(vm *goja.Runtime) {
		defer scriptWg.Done()

		for _, module := range e.modules {
			module(vm, e)
		}

		_, err := vm.RunString(compiledCode)
		if err != nil {
			if jsErr, ok := err.(*goja.Exception); ok {
				if s, ok := any(jsErr).(interface{ Stack() string }); ok {
					runErr = fmt.Errorf("Runtime Error:\n%s", s.Stack())
				} else {
					runErr = fmt.Errorf("Runtime Error:\n%v", jsErr.String())
				}
			} else {
				runErr = fmt.Errorf("Internal Error:\n%v", err)
			}
			return
		}
	})

	scriptWg.Wait()

	if runErr != nil {
		return runErr
	}

	e.Wg.Wait()
	return nil
}
