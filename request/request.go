package request

import (
	"net/http"
)

type HttpRequest struct {
	R        *http.Request
	pathVars map[string]string
}

func (r *HttpRequest) PathVar(varName string) string {
	if r.pathVars != nil {
		return r.pathVars[varName]
	}
	return ""
}

func NewHttpRequest(r *http.Request) *HttpRequest {
	return &HttpRequest{
		R: r,
	}
}

func NewHttpRequestWithPathVars(r *http.Request, uriVars map[string]string) *HttpRequest {
	return &HttpRequest{
		R:        r,
		pathVars: uriVars,
	}
}
