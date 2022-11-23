package response

import (
	"github.com/ixtendio/gofre/router/path"
	"net/http"
)

// HttpHandlerAdaptorResponse adapts the http.HandlerFunc and http.Handler to a response.HttpResponse
type HttpHandlerAdaptorResponse struct {
	headers HttpHeaders
	cookies HttpCookies
	handler http.Handler
}

func (r *HttpHandlerAdaptorResponse) StatusCode() int {
	return 0
}

func (r *HttpHandlerAdaptorResponse) Headers() HttpHeaders {
	if r.headers == nil {
		r.headers = HttpHeaders{}
	}
	return r.headers
}

func (r *HttpHandlerAdaptorResponse) Cookies() HttpCookies {
	if r.cookies == nil {
		r.cookies = HttpCookies{}
	}
	return r.cookies
}

func (r *HttpHandlerAdaptorResponse) Write(w http.ResponseWriter, mc path.MatchingContext) error {
	// Write custom cookies
	if r.cookies != nil {
		for _, cookie := range r.cookies {
			http.SetCookie(w, cookie)
		}
	}

	headers := w.Header()
	// Write custom headers
	if r.headers != nil {
		for k, v := range r.headers {
			headers.Set(k, v)
		}
	}

	r.handler.ServeHTTP(w, mc.R)
	return nil
}

// HandlerFuncAdaptor adapt a http.HandlerFunc to HttpResponse
func HandlerFuncAdaptor(handlerFunc http.HandlerFunc) *HttpHandlerAdaptorResponse {
	return &HttpHandlerAdaptorResponse{
		handler: handlerFunc,
	}
}

// HandlerAdaptor adapt a http.Handler to HttpResponse
func HandlerAdaptor(handler http.Handler) *HttpHandlerAdaptorResponse {
	return &HttpHandlerAdaptorResponse{
		handler: handler,
	}
}
