package response

import (
	"fmt"
	"github.com/ixtendio/gofre/request"
	"io"
	"net/http"
)

type HttpStreamResponse struct {
	HttpHeadersResponse
	Reader io.Reader
}

func (r *HttpStreamResponse) Write(w http.ResponseWriter, req *request.HttpRequest) error {
	// write the headers
	if err := r.HttpHeadersResponse.Write(w, req); err != nil {
		return err
	}
	if r.Reader != nil {
		if _, err := io.Copy(w, r.Reader); err != nil {
			return fmt.Errorf("failed transferig the input stream, err: %w", err)
		}
	}
	return nil
}

// StreamHttpResponse creates a 200 success reader response with a specific content type
func StreamHttpResponse(reader io.Reader, contentType string) *HttpStreamResponse {
	return StreamHttpResponseWithHeadersAndCookies(http.StatusOK, reader, contentType, nil, nil)
}

// StreamHttpResponseWithHeaders creates a 200 success reader response with custom headers
func StreamHttpResponseWithHeaders(statusCode int, reader io.Reader, contentType string, headers http.Header) *HttpStreamResponse {
	return StreamHttpResponseWithHeadersAndCookies(statusCode, reader, contentType, headers, nil)
}

// StreamHttpResponseWithCookies creates a 200 success reader response with custom cookies
func StreamHttpResponseWithCookies(statusCode int, reader io.Reader, contentType string, cookies []http.Cookie) *HttpStreamResponse {
	return StreamHttpResponseWithHeadersAndCookies(statusCode, reader, contentType, nil, cookies)
}

// StreamHttpResponseWithHeadersAndCookies creates a 200 success reader response with custom headers and cookies
func StreamHttpResponseWithHeadersAndCookies(statusCode int, reader io.Reader, contentType string, headers http.Header, cookies []http.Cookie) *HttpStreamResponse {
	if headers == nil {
		headers = http.Header{}
	}
	headers.Set("Content-Type", contentType)
	return &HttpStreamResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    headers,
			HttpCookies:    NewHttpCookies(cookies),
		},
		Reader: reader,
	}
}
