package response

import (
	"github.com/ixtendio/gofre/request"
	"net/http"
)

type HttpHandlerAdaptorResponse struct {
	handler     http.Handler
	handlerFunc http.HandlerFunc
}

func (r *HttpHandlerAdaptorResponse) StatusCode() int {
	return 0
}

func (r *HttpHandlerAdaptorResponse) Headers() http.Header {
	return nil
}

func (r *HttpHandlerAdaptorResponse) Cookies() []*http.Cookie {
	return nil
}

func (r *HttpHandlerAdaptorResponse) Write(w http.ResponseWriter, responseContext *request.HttpRequest) error {
	if r.handler != nil {
		r.handler.ServeHTTP(w, responseContext.R)
	} else {
		r.handlerFunc(w, responseContext.R)
	}
	return nil
}

// HandlerFuncAdaptor adapt a http.HandlerFunc to HttpResponse
func HandlerFuncAdaptor(handlerFunc http.HandlerFunc) *HttpHandlerAdaptorResponse {
	return &HttpHandlerAdaptorResponse{
		handlerFunc: handlerFunc,
	}
}

// HandlerAdaptor adapt a http.Handler to HttpResponse
func HandlerAdaptor(handler http.Handler) *HttpHandlerAdaptorResponse {
	return &HttpHandlerAdaptorResponse{
		handler: handler,
	}
}
