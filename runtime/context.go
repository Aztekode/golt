package runtime

import (
	"encoding/json"
	"io"
	"net/http"
)

type HttpContext struct {
	w      http.ResponseWriter
	r      *http.Request
	status int
}

func (c *HttpContext) Method() string { return c.r.Method }
func (c *HttpContext) Url() string    { return c.r.URL.Path }

func (c *HttpContext) Param(name string) string {
	return c.r.PathValue(name)
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

	var data map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		c.Status(http.StatusBadRequest).Json(map[string]interface{}{
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
		c.Status(http.StatusBadRequest).Json(map[string]interface{}{
			"error":   "Bad request",
			"message": "Invalid body format",
			"errors":  errors,
		})
		return nil
	}

	return data
}
