package runtime

import (
	"fmt"

	"github.com/dop251/goja"
)

func InitConsole(vm *goja.Runtime, e *GoltEngine) {
	console := vm.NewObject()

	console.Set("log", func(call goja.FunctionCall) goja.Value {
		var args []any

		for _, arg := range call.Arguments {
			args = append(args, arg.Export())
		}

		fmt.Println(args...)
		return goja.Undefined()
	})

	vm.Set("console", console)
}
