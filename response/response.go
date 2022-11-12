package response

import (
	"github.com/ixtendio/gofre/request"
	"net/http"
)

type HttpCookies map[string]http.Cookie

func (c HttpCookies) Add(cookie http.Cookie) {
	id := cookie.Name + ":" + cookie.Path + ":" + cookie.Domain
	c[id] = cookie
}

func NewHttpCookies(cookiesArray []http.Cookie) HttpCookies {
	if len(cookiesArray) == 0 {
		return nil
	}
	cookies := make(HttpCookies, len(cookiesArray))
	for _, c := range cookiesArray {
		cookies.Add(c)
	}
	return cookies
}

type HttpResponse interface {
	StatusCode() int
	Headers() http.Header
	Cookies() HttpCookies
	Write(w http.ResponseWriter, reqContext request.HttpRequest) error
}
