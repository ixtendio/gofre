package response

import (
	"github.com/ixtendio/gofre/request"
	"net/http"
)

type HttpHandlerAdaptorResponse struct {
	headers http.Header
	cookies HttpCookies
	handler http.Handler
}

func (r *HttpHandlerAdaptorResponse) StatusCode() int {
	return 0
}

func (r *HttpHandlerAdaptorResponse) Headers() http.Header {
	if r.headers == nil {
		r.headers = http.Header{}
	}
	return r.headers
}

func (r *HttpHandlerAdaptorResponse) Cookies() HttpCookies {
	if r.cookies == nil {
		r.cookies = HttpCookies{}
	}
	return r.cookies
}

func (r *HttpHandlerAdaptorResponse) Write(w http.ResponseWriter, responseContext request.HttpRequest) error {
	// Write custom cookies
	if r.cookies != nil {
		for _, cookie := range r.cookies {
			http.SetCookie(w, &cookie)
		}
	}

	// Write custom headers
	if r.headers != nil {
		for k, v := range r.headers {
			for _, e := range v {
				w.Header().Add(k, e)
			}
		}
	}

	r.handler.ServeHTTP(w, responseContext.R)
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
