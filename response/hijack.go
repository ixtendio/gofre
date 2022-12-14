package response

import (
	"bufio"
	"errors"
	"github.com/ixtendio/gofre/router/path"

	"net"
	"net/http"
)

// HttpHijackConnectionResponse implements response.HttpResponse and provides hijacking capability for the current TCP connection
type HttpHijackConnectionResponse struct {
	hjCallbackFunc func(net.Conn, *bufio.ReadWriter, error)
}

// NewHttpHijackConnectionResponse creates a new HttpResponse that hijack the current TCP connection, giving the full access of it to the application
// This response might be useful in websockets implementations, for example
func NewHttpHijackConnectionResponse(hijackCallbackFunc func(net.Conn, *bufio.ReadWriter, error)) *HttpHijackConnectionResponse {
	return &HttpHijackConnectionResponse{hjCallbackFunc: hijackCallbackFunc}
}

func (r *HttpHijackConnectionResponse) StatusCode() int {
	return 0
}

func (r *HttpHijackConnectionResponse) Headers() HttpHeaders {
	return nil
}

func (r *HttpHijackConnectionResponse) Cookies() HttpCookies {
	return nil
}

func (r *HttpHijackConnectionResponse) Write(w http.ResponseWriter, mc path.MatchingContext) error {
	if hj, ok := w.(http.Hijacker); ok {
		r.hjCallbackFunc(hj.Hijack())
	} else {
		r.hjCallbackFunc(nil, nil, errors.New("the current http.ResponseWriter doesn't support hijack functionality"))
	}
	return nil
}
