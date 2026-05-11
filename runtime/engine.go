package runtime

import (
	"fmt"
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
	buildResult := api.Build(api.BuildOptions{
		EntryPoints: []string{filename},
		Bundle:      true,
		Write:       false,
		Format:      api.FormatIIFE,
		Target:      api.ES2015,
		Platform:    api.PlatformNode,
	})

	if len(buildResult.Errors) > 0 {
		return fmt.Errorf("compilation error in %s: %s", filename, buildResult.Errors[0].Text)
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
				runErr = fmt.Errorf("Runtime Error:\n%v", jsErr.String())
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
