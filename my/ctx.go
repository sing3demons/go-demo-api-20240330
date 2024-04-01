package my

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HandlerFunc func(IContext) error

type IContext interface {
	JSON(statusCode int, data any) error
	Status(statusCode int) error
	Request() *http.Request
	Response() http.ResponseWriter
	Session() string
	Param(key string) string
}

type myContext struct {
	w   http.ResponseWriter
	req *http.Request
}

func NewContext(w http.ResponseWriter, r *http.Request) IContext {
	return &myContext{w, r}
}

func (c *myContext) Session() string {
	return c.req.Context().Value(ContextKey(XSession)).(string)
}

func (c *myContext) Status(statusCode int) error {
	c.w.WriteHeader(statusCode)
	return nil
}

func (c *myContext) JSON(statusCode int, data any) error {
	c.w.WriteHeader(statusCode)
	c.req.Header.Set("Content-Type", "application/json")

	return json.NewEncoder(c.w).Encode(data)
}

func (c *myContext) Request() *http.Request {
	return c.req
}

func (c *myContext) Response() http.ResponseWriter {
	return c.w
}

func (c *myContext) Bind(v any) error {
	// defer c.req.Body.Close()
	return json.NewDecoder(c.req.Body).Decode(v)
}

func (c *myContext) Param(key string) string {
	value, ok := c.req.Context().Value(ContextKey(fmt.Sprintf("{%s}", key))).(string)
	if !ok {
		return ""
	}

	return value
}

func (c *myContext) Query(key string) string {
	return c.req.URL.Query().Get(key)
}

func (c *myContext) Header(key string) string {
	return c.req.Header.Get(key)
}

func (c *myContext) SetHeader(key, value string) {
	c.req.Header.Set(key, value)
}
