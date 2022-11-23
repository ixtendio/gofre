package response

import (
	"fmt"
	"github.com/ixtendio/gofre/router/path"

	"io"
	"net/http"
)

type RawWriterFunc func(w io.Writer) error

// HttpRawResponse implements response.HttpResponse and exposes the http.ResponseWriter for custom write
type HttpRawResponse struct {
	HttpHeadersResponse
	WriteFunc RawWriterFunc
}

func (r *HttpRawResponse) Write(w http.ResponseWriter, mc path.MatchingContext) error {
	// write the headers
	if err := r.HttpHeadersResponse.Write(w, mc); err != nil {
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
// The headers, if present, once will be written to output will be added in the pool for re-use
func RawWriterHttpResponseWithHeaders(statusCode int, contentType string, headers HttpHeaders, writeFunc RawWriterFunc) *HttpRawResponse {
	return RawWriterHttpResponseWithHeadersAndCookies(statusCode, contentType, headers, nil, writeFunc)
}

// RawWriterHttpResponseWithCookies creates a 200 success reader response with custom cookies
// The cookies, if present, once will be written to output will be added in the pool for re-use
func RawWriterHttpResponseWithCookies(statusCode int, contentType string, cookies HttpCookies, writeFunc RawWriterFunc) *HttpRawResponse {
	return RawWriterHttpResponseWithHeadersAndCookies(statusCode, contentType, nil, cookies, writeFunc)
}

// RawWriterHttpResponseWithHeadersAndCookies creates a 200 success reader response with custom headers and cookies
// The headers and cookies, if present, once will be written to output will be added in the pool for re-use
func RawWriterHttpResponseWithHeadersAndCookies(statusCode int, contentType string, headers HttpHeaders, cookies HttpCookies, writeFunc RawWriterFunc) *HttpRawResponse {
	return &HttpRawResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			ContentType:    contentType,
			HttpHeaders:    headers,
			HttpCookies:    cookies,
		},
		WriteFunc: writeFunc,
	}
}
