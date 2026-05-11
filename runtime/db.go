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
	goltObj := vm.Get("Golt").ToObject(vm)
	dbObj := vm.NewObject()

	var activeDB *sql.DB

	createQueryFunc := func(getDB func() *sql.DB) func(goja.FunctionCall) goja.Value {
		return func(call goja.FunctionCall) goja.Value {
			db := getDB()
			if db == nil {
				panic("Database not connected. Call Golt.db.connect() first.")
			}

			querySql := call.Argument(0).String()
			var params []any
			for i := 1; i < len(call.Arguments); i++ {
				params = append(params, call.Argument(i).Export())
			}

			promise, resolve, reject := vm.NewPromise()

			go func() {
				rows, err := db.Query(querySql, params...)
				if err != nil {
					e.loop.RunOnLoop(func(r *goja.Runtime) { reject(vm.ToValue(err)) })
					return
				}
				defer rows.Close()

				columns, err := rows.Columns()
				if err != nil {
					e.loop.RunOnLoop(func(r *goja.Runtime) { reject(vm.ToValue(err)) })
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
						e.loop.RunOnLoop(func(r *goja.Runtime) { reject(vm.ToValue(err)) })
						return
					}

					rowMap := make(map[string]any)
					for i, col := range columns {
						val := values[i]
						if b, ok := val.([]byte); ok {
							rowMap[col] = string(b)
						} else {
							rowMap[col] = val
						}
					}
					results = append(results, rowMap)
				}

				e.loop.RunOnLoop(func(r *goja.Runtime) { resolve(results) })
			}()

			return vm.ToValue(promise)
		}
	}

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
		fmt.Printf("[Golt] * Connected to %s\n", dialect)

		clientObj := vm.NewObject()

		clientObj.Set("query", createQueryFunc(func() *sql.DB { return db }))

		clientObj.Set("close", func(call goja.FunctionCall) goja.Value {
			db.Close()
			return goja.Undefined()
		})

		return clientObj
	})

	dbObj.Set("query", createQueryFunc(func() *sql.DB { return activeDB }))

	goltObj.Set("db", dbObj)
}
