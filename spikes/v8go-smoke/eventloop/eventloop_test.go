//go:build !windows

package eventloop

import (
	"fmt"
	"testing"
	"time"

	v8 "github.com/the-btfash-foundation/v8go"
)

func TestTickerAndGoroutinesFeedV8(t *testing.T) {
	iso := v8.NewIsolate()
	defer iso.Dispose()

	ctx := v8.NewContext(iso)
	defer ctx.Close()

	if _, err := ctx.RunScript(`
		globalThis.__ticks = [];
		globalThis.__promiseValues = [];
		globalThis.__done = false;
		globalThis.onTick = (value) => {
			globalThis.__ticks.push(value);
			Promise.resolve(value * 2).then((resolved) => {
				globalThis.__promiseValues.push(resolved);
				if (globalThis.__promiseValues.length === 3) {
					globalThis.__done = true;
				}
			});
		};
	`, "bootstrap.js"); err != nil {
		t.Fatalf("bootstrap event loop script: %v", err)
	}

	tasks := make(chan func() error, 8)
	producerDone := make(chan struct{})

	go func() {
		defer close(producerDone)

		for i := 1; i <= 3; i++ {
			value := i
			tasks <- func() error {
				_, err := ctx.RunScript(fmt.Sprintf("onTick(%d)", value), fmt.Sprintf("tick-%d.js", value))
				return err
			}

			time.Sleep(5 * time.Millisecond)
		}
	}()

	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.NewTimer(500 * time.Millisecond)
	defer timeout.Stop()

	producerDoneCh := producerDone
	processedTasks := 0

	for {
		select {
		case task := <-tasks:
			if err := task(); err != nil {
				t.Fatalf("run queued task: %v", err)
			}
			processedTasks++
		case <-producerDoneCh:
			producerDoneCh = nil
		case <-ticker.C:
			ctx.PerformMicrotaskCheckpoint()

			if producerDoneCh == nil && processedTasks == 3 {
				done, err := evalString(ctx, "__done ? 'done' : 'pending'", "done.js")
				if err != nil {
					t.Fatalf("read done flag: %v", err)
				}

				if done == "done" {
					snapshot, err := evalString(ctx, "JSON.stringify({ticks: __ticks, promiseValues: __promiseValues, done: __done})", "snapshot.js")
					if err != nil {
						t.Fatalf("read final snapshot: %v", err)
					}

					const want = "{\"ticks\":[1,2,3],\"promiseValues\":[2,4,6],\"done\":true}"
					if snapshot != want {
						t.Fatalf("unexpected event loop snapshot: got %s want %s", snapshot, want)
					}

					return
				}
			}
		case <-timeout.C:
			t.Fatal("timed out waiting for Go ticker/goroutine loop to flush V8 microtasks")
		}
	}
}

func evalString(ctx *v8.Context, script string, filename string) (string, error) {
	value, err := ctx.RunScript(script, filename)
	if err != nil {
		return "", err
	}

	return value.String(), nil
}
