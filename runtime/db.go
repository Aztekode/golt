package runtime

import (
	"database/sql"
	"fmt"

	"github.com/dop251/goja"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"
	_ "modernc.org/sqlite"
)

func InitDB(vm *goja.Runtime, e *GoltEngine) {
	glotObj := vm.Get("Golt").ToObject(vm)
	dbObj := vm.NewObject()

	var activeDB *sql.DB

	dbObj.Set("connect", func(call goja.FunctionCall) goja.Value {
		dialect := call.Argument(0).String()
		dsn := call.Argument(1).String()

		db, err := sql.Open(dialect, dsn)

		if err != nil {
			panic(fmt.Sprintf("Error on connect: %v", err))
		}

		if err := db.Ping(); err != nil {
			panic(fmt.Sprintf("Error on ping: %v", err))
		}

		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(25)
		activeDB = db
		fmt.Printf("\033[36;1m[Golt]: Connected to %s\033[0m\n", dialect)
		return goja.Undefined()
	})

	dbObj.Set("query", func(call goja.FunctionCall) goja.Value {
		if activeDB == nil {
			panic("Database not connected. Call a Golt.db.connect(dialect, url) first.")
		}
		querySql := call.Argument(0).String()
		var params []any
		for i := 1; i < len(call.Arguments); i++ {
			params = append(params, call.Argument(i).Export())
		}

		promise, resolve, reject := vm.NewPromise()

		go func() {
			rows, err := activeDB.Query(querySql, params...)
			if err != nil {
				e.loop.RunOnLoop(func(r *goja.Runtime) {
					reject(vm.ToValue(err))
				})
				return
			}
			defer rows.Close()

			columns, err := rows.Columns()
			if err != nil {
				e.loop.RunOnLoop(func(r *goja.Runtime) {
					reject(vm.ToValue(err))
				})
				return
			}

			count := len(columns)
			var results []map[string]any

			for rows.Next() {
				values := make([]any, count)
				valuePtrs := make([]any, count)
				for i := 0; i < count; i++ {
					valuePtrs[i] = &values[i]
				}

				if err := rows.Scan(valuePtrs...); err != nil {
					e.loop.RunOnLoop(func(r *goja.Runtime) {
						reject(vm.ToValue(err))
					})
					return
				}
				rowMap := make(map[string]any)
				for i, col := range columns {
					val := values[i]

					b, ok := val.([]byte)

					if ok {
						rowMap[col] = string(b)
					} else {
						rowMap[col] = val
					}
				}
				results = append(results, rowMap)

			}
			e.loop.RunOnLoop(func(r *goja.Runtime) {
				resolve(results)
			})
		}()

		return vm.ToValue(promise)
	})

	glotObj.Set("db", dbObj)
}
