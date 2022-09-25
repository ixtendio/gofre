package response

import (
	"github.com/ixtendio/gow/request"
	"net/http"
)

type HttpResponse interface {
	StatusCode() int
	Headers() map[string]string
	Cookies() []*http.Cookie
	Write(w http.ResponseWriter, reqContext *request.HttpRequest) error
}

type HttpHeadersResponse struct {
	HttpStatusCode int
	HttpHeaders    map[string]string
	HttpCookies    []*http.Cookie
}

func (r *HttpHeadersResponse) StatusCode() int {
	if r.HttpStatusCode == 0 {
		return http.StatusOK
	}
	return r.HttpStatusCode
}

func (r *HttpHeadersResponse) Headers() map[string]string {
	return r.HttpHeaders
}

func (r *HttpHeadersResponse) Cookies() []*http.Cookie {
	return r.HttpCookies
}

func (r *HttpHeadersResponse) Write(w http.ResponseWriter, req *request.HttpRequest) error {
	// Write the cookies
	for i := 0; i < len(r.HttpCookies); i++ {
		if r.HttpCookies[i] != nil {
			http.SetCookie(w, r.HttpCookies[i])
		}
	}

	// Write the headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if r.HttpHeaders != nil {
		for k, v := range r.HttpHeaders {
			w.Header().Set(k, v)
		}
	}

	if r.HttpStatusCode > 0 {
		// Write the status code to the response.
		w.WriteHeader(r.HttpStatusCode)
	}
	return nil
}

func InternalServerErrorHttpResponse() *HttpHeadersResponse {
	return &HttpHeadersResponse{
		HttpStatusCode: http.StatusInternalServerError,
	}
}
