package main

import (
	"fmt"

	v8 "github.com/the-btfash-foundation/v8go"
)

func main() {
	iso := v8.NewIsolate()
	defer iso.Dispose()

	ctx := v8.NewContext(iso)
	defer ctx.Close()

	global := ctx.Global()
	if err := global.Set("answer", 41); err != nil {
		panic(fmt.Errorf("set global: %w", err))
	}

	val, err := ctx.RunScript("answer + 1", "math.js")
	if err != nil {
		panic(fmt.Errorf("run math script: %w", err))
	}

	if _, err := ctx.RunScript(`
		globalThis.__promiseResult = 0;
		Promise.resolve(7).then((value) => {
			globalThis.__promiseResult = value * 6;
		});
	`, "promise.js"); err != nil {
		panic(fmt.Errorf("run promise script: %w", err))
	}

	ctx.PerformMicrotaskCheckpoint()

	promiseVal, err := ctx.RunScript("__promiseResult", "result.js")
	if err != nil {
		panic(fmt.Errorf("read promise result: %w", err))
	}

	fmt.Printf("eval=%s promise=%s\n", val, promiseVal)
}
