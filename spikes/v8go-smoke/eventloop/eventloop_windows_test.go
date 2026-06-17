//go:build windows

package eventloop

import "testing"

func TestTickerAndGoroutinesFeedV8(t *testing.T) {
	t.Skip("v8go event loop spike is not executable on Windows until the linker issue is resolved")
}
