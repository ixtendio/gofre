package response

import (
	"errors"
	"github.com/ixtendio/gofre/router/path"
	"net/http"
)

// HttpHeadersResponse implements response.HttpResponse and provides HTTP headers write
type HttpHeadersResponse struct {
	HttpStatusCode int
	ContentType    string
	HttpHeaders    HttpHeaders
	HttpCookies    HttpCookies
}

func (r *HttpHeadersResponse) StatusCode() int {
	if r.HttpStatusCode == 0 {
		return http.StatusOK
	}
	return r.HttpStatusCode
}

func (r *HttpHeadersResponse) Headers() HttpHeaders {
	if r.HttpHeaders == nil {
		r.HttpHeaders = NewHttpHeaders()
	}
	return r.HttpHeaders
}

func (r *HttpHeadersResponse) Cookies() HttpCookies {
	if r.HttpCookies == nil {
		r.HttpCookies = NewEmptyHttpCookie()
	}
	return r.HttpCookies
}

func (r *HttpHeadersResponse) Write(w http.ResponseWriter, mc path.MatchingContext) error {
	defer func() {
		if r.HttpHeaders != nil {
			r.HttpHeaders.Release()
		}
		if r.HttpCookies != nil {
			r.HttpCookies.Release()
		}
	}()

	statusCode := r.StatusCode()
	if statusCode < 100 || statusCode > 999 {
		return errors.New("http status code should be between 100 and 999")
	}

	// Write the cookies
	if r.HttpCookies != nil {
		for _, cookie := range r.HttpCookies {
			http.SetCookie(w, cookie)
		}
	}

	header := w.Header()

	// Write the headers
	if r.HttpHeaders != nil {
		for k, v := range r.HttpHeaders {
			header.Set(k, v)
		}
	}

	if len(header.Get(HeaderContentType)) == 0 && len(r.ContentType) > 0 {
		header.Set(HeaderContentType, r.ContentType)
	}
	if len(header.Get(HeaderContentTypeOptions)) == 0 {
		header.Set(HeaderContentTypeOptions, "nosniff")
	}

	// Write the status code to the response.
	w.WriteHeader(statusCode)
	return nil
}

// InternalServerErrorHttpResponse writes the http.StatusInternalServerError HTTP status code to the client
func InternalServerErrorHttpResponse() *HttpHeadersResponse {
	return &HttpHeadersResponse{
		HttpStatusCode: http.StatusInternalServerError,
	}
}
