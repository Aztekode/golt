package runtime

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/dop251/goja"
)

func InitFetch(vm *goja.Runtime, e *GoltEngine) {
	vm.Set("fetch", func(call goja.FunctionCall) goja.Value {
		promise, resolve, reject := vm.NewPromise()

		if len(call.Arguments) == 0 {
			reject(vm.ToValue("TypeError: fetch requires at least 1 argument"))
			return vm.ToValue(promise)
		}

		url := call.Argument(0).String()
		method := "GET"
		var bodyBytes []byte
		headers := make(map[string]string)

		timeoutDuration := 15 * time.Second

		if len(call.Arguments) > 1 {
			opts := call.Argument(1).ToObject(vm)

			if m := opts.Get("method"); m != nil && !goja.IsUndefined(m) {
				method = m.String()
			}

			if h := opts.Get("headers"); h != nil && !goja.IsUndefined(h) {
				hObj := h.ToObject(vm)
				for _, key := range hObj.Keys() {
					headers[key] = hObj.Get(key).String()
				}
			}

			if b := opts.Get("body"); b != nil && !goja.IsUndefined(b) && !goja.IsNull(b) {
				bodyBytes = []byte(b.String())
			}

			if t := opts.Get("timeout"); t != nil && !goja.IsUndefined(t) {
				timeoutDuration = time.Duration(t.ToInteger()) * time.Millisecond
			}
		}

		go func() {
			req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
			if err != nil {
				e.loop.RunOnLoop(func(*goja.Runtime) { reject(err.Error()) })
				return
			}

			for k, v := range headers {
				req.Header.Set(k, v)
			}

			client := &http.Client{
				Timeout: timeoutDuration,
			}

			resp, err := client.Do(req)
			if err != nil {
				e.loop.RunOnLoop(func(*goja.Runtime) { reject(err.Error()) })
				return
			}
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				e.loop.RunOnLoop(func(*goja.Runtime) { reject(err.Error()) })
				return
			}

			e.loop.RunOnLoop(func(*goja.Runtime) {
				respObj := vm.NewObject()
				respObj.Set("ok", resp.StatusCode >= 200 && resp.StatusCode < 300)
				respObj.Set("status", resp.StatusCode)

				respObj.Set("statusText", http.StatusText(resp.StatusCode))

				headersObj := vm.NewObject()
				headersObj.Set("get", func(c goja.FunctionCall) goja.Value {
					key := c.Argument(0).String()
					return vm.ToValue(resp.Header.Get(key))
				})
				respObj.Set("headers", headersObj)

				respObj.Set("text", func(goja.FunctionCall) goja.Value {
					p, res, _ := vm.NewPromise()
					res(string(respBody))
					return vm.ToValue(p)
				})

				respObj.Set("json", func(goja.FunctionCall) goja.Value {
					p, res, rej := vm.NewPromise()
					var jsonData interface{}
					if err := json.Unmarshal(respBody, &jsonData); err != nil {
						rej(err.Error())
					} else {
						res(jsonData)
					}
					return vm.ToValue(p)
				})

				resolve(respObj)
			})
		}()

		return vm.ToValue(promise)
	})
}
