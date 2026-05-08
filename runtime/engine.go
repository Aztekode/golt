package runtime

import (
	"fmt"
	"os"
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

func (e *GoltEngine) RunFile(filename string) {
	buildResult := api.Build(api.BuildOptions{
		EntryPoints: []string{filename},
		Bundle:      true,
		Write:       false,
		Format:      api.FormatIIFE,
		Target:      api.ES2015,
		Platform:    api.PlatformNode,
	})

	if len(buildResult.Errors) > 0 {
		fmt.Printf("Error on compilation: %s:\n", filename)
		for _, err := range buildResult.Errors {
			fmt.Printf("- %s\n", err.Text)
			os.Exit(1)
		}
	}

	compiledCode := string(buildResult.OutputFiles[0].Contents)

	e.loop.Start()
	defer e.loop.Stop()

	var scriptWg sync.WaitGroup
	scriptWg.Add(1)

	e.loop.RunOnLoop(func(vm *goja.Runtime) {
		for _, module := range e.modules {
			module(vm, e)
		}

		_, err := vm.RunString(compiledCode)
		if err != nil {
			if jsErr, ok := err.(*goja.Exception); ok {
				fmt.Printf("Error on Golt Runtime:\n%v\n", jsErr.String())
			} else {
				fmt.Printf("Internal Error:\n%v\n", err)
			}
			os.Exit(1)
		}

		scriptWg.Done()
	})

	scriptWg.Wait()
	e.Wg.Wait()
}
