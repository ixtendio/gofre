package request

import (
	"github.com/ixtendio/gofre/internal/path"
	"net/http"
)

type HttpRequest struct {
	R        *http.Request
	pathVars []path.CaptureVar
}

func (r *HttpRequest) PathVar(varName string) string {
	for i := 0; i < len(r.pathVars); i++ {
		pv := &r.pathVars[i]
		if pv.Name == varName {
			return pv.Value
		}
	}
	return ""
}

func (r *HttpRequest) PathVars() []path.CaptureVar {
	return r.pathVars
}

func NewHttpRequest(r *http.Request) HttpRequest {
	return HttpRequest{
		R: r,
	}
}

func NewHttpRequestWithPathVars(r *http.Request, pathVars []path.CaptureVar) HttpRequest {
	return HttpRequest{
		R:        r,
		pathVars: pathVars,
	}
}
