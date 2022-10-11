package response

import (
	"github.com/ixtendio/gofre/request"
	"net/http"
)

type HttpResponse interface {
	StatusCode() int
	Headers() http.Header
	Cookies() []*http.Cookie
	Write(w http.ResponseWriter, reqContext *request.HttpRequest) error
}

type HttpHeadersResponse struct {
	HttpStatusCode int
	HttpHeaders    http.Header
	HttpCookies    []*http.Cookie
}

func (r *HttpHeadersResponse) StatusCode() int {
	if r.HttpStatusCode == 0 {
		return http.StatusOK
	}
	return r.HttpStatusCode
}

func (r *HttpHeadersResponse) Headers() http.Header {
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
			for _, e := range v {
				w.Header().Add(k, e)
			}
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
