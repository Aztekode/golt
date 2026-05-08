package runtime

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/dop251/goja"
)

func InitLogger(vm *goja.Runtime, e *GoltEngine) {
	jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	goltObj := vm.Get("Golt").ToObject(vm)

	goltObj.Set("logger", func(call goja.FunctionCall) goja.Value {
		format := "dev"

		if len(call.Arguments) > 0 {
			configObj := call.Argument(0).ToObject(vm)
			if fmtVal := configObj.Get("format"); fmtVal != nil {
				format = fmtVal.String()
			}
		}

		middleware := func(middlewareCall goja.FunctionCall) goja.Value {
			ctxVal := middlewareCall.Argument(0)
			nextVal := middlewareCall.Argument(1)

			ctx, ok := ctxVal.Export().(*HttpContext)
			if !ok {
				panic("Logger: unable to cast HttpContext")
			}

			start := time.Now()

			nextFn, ok := goja.AssertFunction(nextVal)
			if ok {
				nextFn(goja.Undefined())
			}

			duration := time.Since(start)
			status := ctx.status
			if status == 0 {
				status = 200
			}

			var timeStr string
			var timeColor string

			if duration >= time.Millisecond {
				timeStr = fmt.Sprintf("%.2fms", float64(duration)/float64(time.Millisecond))
			} else if duration >= time.Microsecond {
				timeStr = fmt.Sprintf("%.2fµs", float64(duration)/float64(time.Microsecond))
			} else if duration > 0 {
				timeStr = fmt.Sprintf("%dns", duration.Nanoseconds())
			} else {
				timeStr = "<1µs"
			}

			if duration >= 500*time.Millisecond {
				timeColor = "\033[31;1m"
			} else if duration >= 100*time.Millisecond {
				timeColor = "\033[33;1m"
			} else {
				timeColor = "\033[2m"
			}

			reset := "\033[0m"
			dim := "\033[2m"

			var methodBadge string
			switch ctx.Method() {
			case "GET":
				methodBadge = fmt.Sprintf("\033[44;37;1m %-6s %s", ctx.Method(), reset)
			case "POST":
				methodBadge = fmt.Sprintf("\033[42;37;1m %-6s %s", ctx.Method(), reset)
			case "PUT", "PATCH":
				methodBadge = fmt.Sprintf("\033[43;30;1m %-6s %s", ctx.Method(), reset)
			case "DELETE":
				methodBadge = fmt.Sprintf("\033[41;37;1m %-6s %s", ctx.Method(), reset)
			default:
				methodBadge = fmt.Sprintf("\033[46;37;1m %-6s %s", ctx.Method(), reset)
			}

			var statusBadge string
			switch {
			case status >= 200 && status < 300:
				statusBadge = fmt.Sprintf("\033[42;37;1m %3d %s", status, reset) // Verde
			case status >= 300 && status < 400:
				statusBadge = fmt.Sprintf("\033[46;37;1m %3d %s", status, reset) // Cyan
			case status >= 400 && status < 500:
				statusBadge = fmt.Sprintf("\033[43;30;1m %3d %s", status, reset) // Amarillo
			case status >= 500:
				statusBadge = fmt.Sprintf("\033[41;37;1m %3d %s", status, reset) // Rojo
			default:
				statusBadge = fmt.Sprintf("\033[100;37;1m %3d %s", status, reset) // Gris
			}

			switch format {
			case "json":
				jsonLogger.Info("Request",
					slog.String("method", ctx.Method()),
					slog.String("url", ctx.Url()),
					slog.Int("status", status),
					slog.String("latency", duration.String()),
				)
			case "tiny":
				fmt.Printf("%s %s %d - %s\n", ctx.Method(), ctx.Url(), status, timeStr)
			default: // "dev"
				fmt.Printf(" %s %-30s %s %s %s%10s%s\n",
					methodBadge,
					ctx.Url(),
					statusBadge,
					dim+"|"+reset,
					timeColor, timeStr, reset,
				)
			}

			return goja.Undefined()
		}

		return vm.ToValue(middleware)
	})
}
