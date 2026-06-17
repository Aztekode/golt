package bench

import (
	"testing"

	"github.com/dop251/goja"
)

func BenchmarkColdStartGoja(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		vm := goja.New()
		value, err := vm.RunString(coldStartScript)
		if err != nil {
			b.Fatalf("run goja script: %v", err)
		}

		if value.String() != expectedResult {
			b.Fatalf("unexpected goja result: got %s want %s", value.String(), expectedResult)
		}
	}
}
