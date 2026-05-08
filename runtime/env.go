package runtime

import (
	"os"
	"strings"

	"github.com/dop251/goja"
)

func InitEnv(vm *goja.Runtime, e *GoltEngine) {
	goltObj := vm.Get("Golt")
	var golt *goja.Object

	if goltObj == nil {
		golt = vm.NewObject()
		vm.Set("Golt", golt)
	} else {
		golt = goltObj.ToObject(vm)
	}

	sysEnvs := os.Environ()
	envMap := make(map[string]string)

	for _, env := range sysEnvs {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
		golt.Set("env", envMap)
	}
}
