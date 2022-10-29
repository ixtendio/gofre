package response

import (
	"fmt"
	"github.com/ixtendio/gofre/request"
	"io"
	"net/http"
)

type RawWriterFunc func(w io.Writer) error

type HttpRawResponse struct {
	HttpHeadersResponse
	WriteFunc RawWriterFunc
}

func (r *HttpRawResponse) Write(w http.ResponseWriter, req *request.HttpRequest) error {
	// write the headers
	if err := r.HttpHeadersResponse.Write(w, req); err != nil {
		return err
	}
	if r.WriteFunc != nil {
		if err := r.WriteFunc(w); err != nil {
			return fmt.Errorf("failed to write raw data, err: %w", err)
		}
	}
	return nil
}

// RawWriterHttpResponse creates a 200 success reader response with a specific content type
func RawWriterHttpResponse(contentType string, writeFunc RawWriterFunc) *HttpRawResponse {
	return RawWriterHttpResponseWithHeadersAndCookies(http.StatusOK, contentType, nil, nil, writeFunc)
}

// RawWriterHttpResponseWithHeaders creates a 200 success reader response with custom headers
func RawWriterHttpResponseWithHeaders(statusCode int, contentType string, headers http.Header, writeFunc RawWriterFunc) *HttpRawResponse {
	return RawWriterHttpResponseWithHeadersAndCookies(statusCode, contentType, headers, nil, writeFunc)
}

// RawWriterHttpResponseWithCookies creates a 200 success reader response with custom cookies
func RawWriterHttpResponseWithCookies(statusCode int, contentType string, cookies []http.Cookie, writeFunc RawWriterFunc) *HttpRawResponse {
	return RawWriterHttpResponseWithHeadersAndCookies(statusCode, contentType, nil, cookies, writeFunc)
}

// RawWriterHttpResponseWithHeadersAndCookies creates a 200 success reader response with custom headers and cookies
func RawWriterHttpResponseWithHeadersAndCookies(statusCode int, contentType string, headers http.Header, cookies []http.Cookie, writeFunc RawWriterFunc) *HttpRawResponse {
	if headers == nil {
		headers = http.Header{}
	}
	headers.Set("Content-Type", contentType)
	return &HttpRawResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    headers,
			HttpCookies:    NewHttpCookies(cookies),
		},
		WriteFunc: writeFunc,
	}
}
