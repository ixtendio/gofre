package response

import (
	"errors"
	"github.com/ixtendio/gofre/request"
	"net/http"
)

type HttpHeadersResponse struct {
	HttpStatusCode int
	HttpHeaders    http.Header
	HttpCookies    HttpCookies
}

func (r *HttpHeadersResponse) StatusCode() int {
	if r.HttpStatusCode == 0 {
		return http.StatusOK
	}
	return r.HttpStatusCode
}

func (r *HttpHeadersResponse) Headers() http.Header {
	if r.HttpHeaders == nil {
		r.HttpHeaders = http.Header{}
	}
	return r.HttpHeaders
}

func (r *HttpHeadersResponse) Cookies() HttpCookies {
	if r.HttpCookies == nil {
		r.HttpCookies = HttpCookies{}
	}
	return r.HttpCookies
}

func (r *HttpHeadersResponse) Write(w http.ResponseWriter, req *request.HttpRequest) error {
	statusCode := r.StatusCode()
	if statusCode < 100 || statusCode > 999 {
		return errors.New("http status code should be between 100 and 999")
	}

	// Write the cookies
	if r.HttpCookies != nil {
		for _, cookie := range r.HttpCookies {
			http.SetCookie(w, &cookie)
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

	// Write the status code to the response.
	w.WriteHeader(statusCode)
	return nil
}

func InternalServerErrorHttpResponse() *HttpHeadersResponse {
	return &HttpHeadersResponse{
		HttpStatusCode: http.StatusInternalServerError,
	}
}
