package runtime

import (
	"fmt"
	"net/http"

	"github.com/dop251/goja"
)

type HttpContext struct {
	w      http.ResponseWriter
	r      *http.Request
	status int
}

func (c *HttpContext) Method() string { return c.r.Method }
func (c *HttpContext) Url() string    { return c.r.URL.Path }

func (c *HttpContext) Status(code int) *HttpContext {
	c.status = code
	return c
}

func (c *HttpContext) Send(body string) {
	if c.status == 0 {
		c.status = http.StatusOK
	}

	c.w.WriteHeader(c.status)
	c.w.Write([]byte(body))
}

func InitHttp(vm *goja.Runtime, e *GoltEngine) {
	goltObj := vm.Get("Golt").ToObject(vm)

	goltObj.Set("serve", func(call goja.FunctionCall) goja.Value {
		port := call.Argument(0).ToInteger()
		handler, ok := goja.AssertFunction(call.Argument(1))

		if !ok {
			panic("Golt.serve require a callback function")
		}

		e.Wg.Add(1)

		go func() {
			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				done := make(chan struct{})

				e.loop.RunOnLoop(func(vm *goja.Runtime) {
					ctx := &HttpContext{w: w, r: r}
					_, err := handler(goja.Undefined(), vm.ToValue(ctx))
					if err != nil {
						fmt.Println("Error HTTP: ", err)
						http.Error(w, "Internal server error", http.StatusInternalServerError)
					}
					close(done)
				})
				<-done
			})

			fmt.Printf("Server is running on port %d\n", port)
			if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
				fmt.Println("Error HTTP: ", err)
			}
		}()

		return goja.Undefined()
	})
}
