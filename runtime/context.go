package runtime

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

type HttpContext struct {
	w      http.ResponseWriter
	r      *http.Request
	status int
	done   chan struct{}
	locals map[string]any
	mu     sync.Once
}

func (c *HttpContext) finish() {
	c.mu.Do(func() {
		close(c.done)
	})
}

func (c *HttpContext) Method() string { return c.r.Method }
func (c *HttpContext) Url() string    { return c.r.URL.Path }

func (c *HttpContext) Param(name string) string {
	return c.r.PathValue(name)
}

func (c *HttpContext) Set(key string, value any) {
	c.locals[key] = value
}

func (c *HttpContext) Get(key string) any {
	return c.locals[key]
}

func (c *HttpContext) GetHeader(key string) string {
	return c.r.Header.Get(key)
}

func (c *HttpContext) SetHeader(key, value string) {
	c.w.Header().Set(key, value)
}

func (c *HttpContext) Query(key string) string {
	return c.r.URL.Query().Get(key)
}

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
	c.finish()
}

func (c *HttpContext) Json(data interface{}) {
	if c.status == 0 {
		c.status = http.StatusOK
	}
	c.w.Header().Set("Content-Type", "application/json")
	c.w.WriteHeader(c.status)

	bytes, err := json.Marshal(data)
	if err != nil {
		c.w.Write([]byte(`{ "error": "Internal server error" }`))
		return
	}
	c.w.Write(bytes)
	c.finish()
}

func (c *HttpContext) ValidateBody(schema map[string]interface{}) interface{} {
	bodyBytes, err := io.ReadAll(c.r.Body)
	if err != nil || len(bodyBytes) == 0 {
		c.Status(http.StatusBadRequest).Json(map[string]interface{}{
			"error":   "Bad request",
			"message": "No body provided",
		})
		return nil
	}

	var data map[string]any
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		c.Status(http.StatusBadRequest).Json(map[string]any{
			"error":   "Bad request",
			"message": "Invalid body format",
		})
		return nil
	}

	errors := make(map[string]string)
	for key, expectedType := range schema {
		val, exists := data[key]
		if !exists {
			errors[key] = "required"
			continue
		}

		expStr, ok := expectedType.(string)
		if !ok {
			continue
		}

		switch expStr {
		case "string":
			if _, ok := val.(string); !ok {
				errors[key] = "invalid type: string"
			}
		case "number":
			if _, ok := val.(float64); !ok {
				errors[key] = "invalid type: number"
			}
		case "boolean":
			if _, ok := val.(bool); !ok {
				errors[key] = "invalid type: boolean"
			}
		}
	}

	if len(errors) > 0 {
		c.Status(http.StatusBadRequest).Json(map[string]any{
			"error":   "Bad request",
			"message": "Invalid body format",
			"errors":  errors,
		})
		return nil
	}

	return data
}
