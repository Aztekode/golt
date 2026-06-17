//go:build !windows

package bench

import (
	"testing"

	v8 "github.com/the-btfash-foundation/v8go"
)

func BenchmarkColdStartV8Go(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		iso := v8.NewIsolate()
		ctx := v8.NewContext(iso)

		value, err := ctx.RunScript(coldStartScript, "cold-start.js")
		if err != nil {
			ctx.Close()
			iso.Dispose()
			b.Fatalf("run v8go script: %v", err)
		}

		if value.String() != expectedResult {
			ctx.Close()
			iso.Dispose()
			b.Fatalf("unexpected v8go result: got %s want %s", value.String(), expectedResult)
		}

		ctx.Close()
		iso.Dispose()
	}
}
