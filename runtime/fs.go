package runtime

import (
	"fmt"
	"os"

	"github.com/dop251/goja"
)

func InitFs(vm *goja.Runtime, e *GoltEngine) {
	goltObj := vm.Get("Golt").ToObject(vm)
	fsObj := vm.NewObject()

	fsObj.Set("readFile", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		content, err := os.ReadFile(path)
		if err != nil {
			panic(fmt.Sprintf("FS Error: Could not read file %s", path))
		}
		return vm.ToValue(string(content))
	})

	fsObj.Set("writeFile", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		content := call.Argument(1).String()
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			panic(fmt.Sprintf("FS Error: Could not write file %s", path))
		}
		return goja.Undefined()
	})

	goltObj.Set("fs", fsObj)
}
