package runtime

import (
	"fmt"
	"time"

	"github.com/dop251/goja"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func InitCrypto(vm *goja.Runtime, e *GoltEngine) {
	goltObj := vm.Get("Golt").ToObject(vm)

	cryptoObj := vm.NewObject()

	cryptoObj.Set("hash", func(call goja.FunctionCall) goja.Value {
		password := call.Argument(0).String()
		cost := bcrypt.DefaultCost

		if len(call.Arguments) > 1 {
			cost = int(call.Argument(1).ToInteger())
		}

		promise, resolve, reject := vm.NewPromise()

		go func() {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
			e.RunOnLoop(func(*goja.Runtime) {
				if err != nil {
					reject(err.Error())
				} else {
					resolve(string(hash))
				}
			})
		}()

		return vm.ToValue(promise)
	})

	cryptoObj.Set("compare", func(call goja.FunctionCall) goja.Value {
		password := call.Argument(0).String()
		hash := call.Argument(1).String()

		promise, resolve, reject := vm.NewPromise()

		go func() {
			err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
			e.RunOnLoop(func(*goja.Runtime) {
				if err != nil {
					if err == bcrypt.ErrMismatchedHashAndPassword {
						resolve(false)
					} else {
						reject(err.Error())
					}
				} else {
					resolve(true)
				}
			})
		}()

		return vm.ToValue(promise)
	})

	goltObj.Set("crypto", cryptoObj)

	jwtObj := vm.NewObject()

	jwtObj.Set("sign", func(call goja.FunctionCall) goja.Value {
		payloadObj := call.Argument(0).ToObject(vm)
		secret := []byte(call.Argument(1).String())

		expHours := 24
		if len(call.Arguments) > 2 {
			expHours = int(call.Argument(2).ToInteger())
		}

		claims := jwt.MapClaims{}
		for _, key := range payloadObj.Keys() {
			claims[key] = payloadObj.Get(key).Export()
		}

		claims["exp"] = time.Now().Add(time.Hour * time.Duration(expHours)).Unix()

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedString, err := token.SignedString(secret)
		if err != nil {
			panic(fmt.Sprintf("JWT Sign Error: %v", err))
		}

		return vm.ToValue(signedString)
	})

	jwtObj.Set("verify", func(call goja.FunctionCall) goja.Value {
		tokenString := call.Argument(0).String()
		secret := []byte(call.Argument(1).String())

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method")
			}
			return secret, nil
		})

		if err != nil || !token.Valid {
			return goja.Null()
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			return vm.ToValue(claims)
		}

		return goja.Null()
	})

	goltObj.Set("jwt", jwtObj)
}
