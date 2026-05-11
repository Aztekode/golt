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
				panic("app.use requires a function callback")
			}
			middlewares = append(middlewares, handler)
			return appObj
		})

		registerRoute := func(method string, call goja.FunctionCall) goja.Value {
			path := call.Argument(0).String()
			handler, ok := goja.AssertFunction(call.Argument(1))

			if !ok {
				panic(fmt.Sprintf("app.%s requires a function callback", method))
			}

			exactPath := path
			if exactPath == "/" {
				exactPath = "/{$}"
			}

			routerPattern := fmt.Sprintf("%s %s", method, exactPath)

			mux.HandleFunc(routerPattern, func(w http.ResponseWriter, r *http.Request) {
				ctx := &HttpContext{
					w:      w,
					r:      r,
					done:   make(chan struct{}),
					locals: make(map[string]any),
				}

				e.loop.RunOnLoop(func(vm *goja.Runtime) {
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
								ctx.finish()
							}
						}
					}

					next()
				})

				<-ctx.done
			})
			return appObj
		}

		appObj.Set("get", func(call goja.FunctionCall) goja.Value { return registerRoute("GET", call) })
		appObj.Set("post", func(call goja.FunctionCall) goja.Value { return registerRoute("POST", call) })
		appObj.Set("put", func(call goja.FunctionCall) goja.Value { return registerRoute("PUT", call) })
		appObj.Set("delete", func(call goja.FunctionCall) goja.Value { return registerRoute("DELETE", call) })

		appObj.Set("static", func(call goja.FunctionCall) goja.Value {
			prefix := call.Argument(0).String()
			dirPath := call.Argument(1).String()

			spaFallback := false
			if len(call.Arguments) > 2 {
				spaFallback = call.Argument(2).ToBoolean()
			}

			if prefix == "" || dirPath == "" {
				panic("app.static requires prefix and dirPath arguments")
			}

			pattern := fmt.Sprintf("GET %s/", prefix)

			fileServer := http.StripPrefix(prefix, http.FileServer(http.Dir(dirPath)))

			mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
				if spaFallback {
					path := r.URL.Path[len(prefix):]
					fpath := fmt.Sprintf("%s/%s", dirPath, path)
					if _, err := os.Stat(fpath); os.IsNotExist(err) {
						http.ServeFile(w, r, fmt.Sprintf("%s/index.html", dirPath))
						return
					}
				}
				fileServer.ServeHTTP(w, r)
			})

			fmt.Printf("\033[35;1m[Golt] Static folder mapped: %s -> %s\033[0m\n", prefix, dirPath)

			return appObj
		})

		appObj.Set("notFound", func(call goja.FunctionCall) goja.Value {
			handler, ok := goja.AssertFunction(call.Argument(0))
			if !ok {
				panic("app.notFound requires a function callback")
			}

			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				// Aplicamos exactamente el mismo ajuste para el handler de 404
				ctx := &HttpContext{
					w:    w,
					r:    r,
					done: make(chan struct{}),
				}

				e.loop.RunOnLoop(func(vm *goja.Runtime) {
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
								ctx.finish()
							}
						}
					}
					next()
				})
				<-ctx.done
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
