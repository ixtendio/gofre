package response

import (
	"fmt"
	"github.com/ixtendio/gofre/router/path"
	"io"
	"net/http"
	"reflect"
	"sync"
	"unsafe"
)

const (
	HeaderContentType        = "Content-Type"
	HeaderContentTypeOptions = "X-Content-Type-Options"
)

var httpHeadersPool = sync.Pool{
	New: func() interface{} {
		return make(HttpHeaders)
	},
}

// HttpHeaders is a map with custom headers
type HttpHeaders map[string]string

func (h HttpHeaders) Set(key string, val string) {
	h[key] = val
}

// Clear all the data from the map
func (h HttpHeaders) Clear() {
	for k := range h {
		delete(h, k)
	}
}

// Release put this object in the pool for further re-usage, cleaning it beforehand
func (h HttpHeaders) Release() {
	h.Clear()
	httpHeadersPool.Put(h)
}

// NewHttpHeaders returns an instance of HttpHeaders from the pool
func NewHttpHeaders() HttpHeaders {
	return httpHeadersPool.Get().(HttpHeaders)
}

var httpCookiesPool = sync.Pool{
	New: func() interface{} {
		return make(HttpCookies)
	},
}

// HttpCookies is a map with custom cookies
type HttpCookies map[string]*http.Cookie

func (c HttpCookies) Add(cookies ...*http.Cookie) {
	for i := 0; i < len(cookies); i++ {
		cookie := cookies[i]
		if cookie != nil {
			id := cookie.Name + ":" + cookie.Path + ":" + cookie.Domain
			c[id] = cookie
		}
	}
}

// Clear all the data from the map
func (c HttpCookies) Clear() {
	for k := range c {
		delete(c, k)
	}
}

// Release put this object in the pool for further re-usage, cleaning it beforehand
func (c HttpCookies) Release() {
	c.Clear()
	httpCookiesPool.Put(c)
}

// NewEmptyHttpCookie returns an instance of HttpCookies from the pool
func NewEmptyHttpCookie() HttpCookies {
	return httpCookiesPool.Get().(HttpCookies)
}

// NewHttpCookie returns an instance of HttpCookies from the pool and populates it with the cookies argument
func NewHttpCookie(cookies ...*http.Cookie) HttpCookies {
	c := NewEmptyHttpCookie()
	c.Add(cookies...)
	return c
}

// HttpResponse describes the methods that a custom response should implement
type HttpResponse interface {
	// StatusCode returns the response status code
	StatusCode() int
	// Headers returns the response headers
	Headers() HttpHeaders
	// Cookies returns the response cookies
	Cookies() HttpCookies
	// Write the response to the client
	Write(w http.ResponseWriter, mc path.MatchingContext) error
}

func writeTextResponse(w http.ResponseWriter, payload string) error {
	// write the response
	payloadLen := len(payload)
	if payloadLen > 0 {
		if sw, ok := w.(io.StringWriter); ok {
			if _, err := sw.WriteString(payload); err != nil {
				return fmt.Errorf("failed to write the text response, err: %w", err)
			}
		} else {
			if _, err := w.Write((*[0x7fff0000]byte)(unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&payload)).Data))[:payloadLen:payloadLen]); err != nil {
				return fmt.Errorf("failed to write the text response, err: %w", err)
			}
		}
	}
	return nil
}
