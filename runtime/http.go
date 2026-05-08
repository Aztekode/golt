package runtime

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dop251/goja"
)

func InitHttp(vm *goja.Runtime, e *GoltEngine) {
	goltObj := vm.Get("Golt").ToObject(vm)

	goltObj.Set("App", func(call goja.FunctionCall) goja.Value {
		mux := http.NewServeMux()
		appObj := vm.NewObject()

		var middlewares []goja.Callable

		appObj.Set("use", func(call goja.FunctionCall) goja.Value {
			handler, ok := goja.AssertFunction(call.Argument(0))
			if !ok {
				panic("app.use require a function callback")
			}
			middlewares = append(middlewares, handler)
			return appObj
		})

		registerRoute := func(method string, call goja.FunctionCall) goja.Value {
			path := call.Argument(0).String()
			handler, ok := goja.AssertFunction(call.Argument(1))

			if !ok {
				panic(fmt.Sprintf("app.%s require a function callback", method))
			}

			exactPath := path
			if exactPath == "/" {
				exactPath = "/{$}"
			}

			routerPattern := fmt.Sprintf("%s %s", method, exactPath)

			mux.HandleFunc(routerPattern, func(w http.ResponseWriter, r *http.Request) {
				done := make(chan struct{})
				e.loop.RunOnLoop(func(vm *goja.Runtime) {
					defer close(done)

					ctx := &HttpContext{w: w, r: r}
					ctxVal := vm.ToValue(ctx)

					chain := append([]goja.Callable{}, middlewares...)
					chain = append(chain, handler)
					index := -1

					var next func()
					next = func() {
						index++
						if index < len(chain) {
							_, err := chain[index](goja.Undefined(), ctxVal, vm.ToValue(next))
							if err != nil {
								fmt.Println("Error HTTP: ", err)
								http.Error(w, "Internal server error", http.StatusInternalServerError)
							}
						}
					}

					next()
				})
				<-done
			})
			return appObj
		}

		appObj.Set("get", func(call goja.FunctionCall) goja.Value { return registerRoute("GET", call) })
		appObj.Set("post", func(call goja.FunctionCall) goja.Value { return registerRoute("POST", call) })
		appObj.Set("put", func(call goja.FunctionCall) goja.Value { return registerRoute("PUT", call) })
		appObj.Set("delete", func(call goja.FunctionCall) goja.Value { return registerRoute("DELETE", call) })

		appObj.Set("notFound", func(call goja.FunctionCall) goja.Value {
			handler, ok := goja.AssertFunction(call.Argument(0))
			if !ok {
				panic("app.notFound require a function callback")
			}

			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				done := make(chan struct{})
				e.loop.RunOnLoop(func(vm *goja.Runtime) {
					defer close(done)
					ctx := &HttpContext{w: w, r: r}
					ctx.status = http.StatusNotFound
					ctxVal := vm.ToValue(ctx)

					chain := append([]goja.Callable{}, middlewares...)
					chain = append(chain, handler)
					index := -1

					var next func()
					next = func() {
						index++
						if index < len(chain) {
							_, err := chain[index](goja.Undefined(), ctxVal, vm.ToValue(next))
							if err != nil {
								fmt.Println("Error HTTP: ", err)
								http.Error(w, "Internal server error", http.StatusInternalServerError)
							}
						}
					}
					next()
				})
				<-done
			})
			return appObj
		})

		appObj.Set("serve", func(call goja.FunctionCall) goja.Value {
			port := call.Argument(0).ToInteger()
			e.Wg.Add(1)

			srv := &http.Server{
				Addr:    fmt.Sprintf(":%d", port),
				Handler: mux,
			}

			go func() {
				fmt.Printf("Server is running on port %d\n", port)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					fmt.Println("Critical error on server: ", err)
				}
			}()

			go func() {
				quit := make(chan os.Signal, 1)
				signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
				<-quit

				fmt.Println("Server is shutting down...")
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if err := srv.Shutdown(ctx); err != nil {
					fmt.Println("Error on server shutdown: ", err)
				}

				fmt.Println("Server is down")
				e.loop.Stop()
				e.Wg.Done()
			}()

			return goja.Undefined()
		})

		return appObj
	})
}
