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

func (r *HttpStreamResponse) Write(w http.ResponseWriter, req request.HttpRequest) error {
	// write the headers
	if err := r.HttpHeadersResponse.Write(w, req); err != nil {
		return err
	}
	if r.Reader != nil {
		if _, err := io.Copy(w, r.Reader); err != nil {
			return fmt.Errorf("failed to transfer the input stream, err: %w", err)
		}
	}
	return nil
}

// StreamHttpResponse creates a 200 success reader response with a specific content type
func StreamHttpResponse(contentType string, reader io.Reader) *HttpStreamResponse {
	return StreamHttpResponseWithHeadersAndCookies(http.StatusOK, contentType, nil, nil, reader)
}

// StreamHttpResponseWithHeaders creates a 200 success reader response with custom headers
func StreamHttpResponseWithHeaders(statusCode int, contentType string, headers http.Header, reader io.Reader) *HttpStreamResponse {
	return StreamHttpResponseWithHeadersAndCookies(statusCode, contentType, headers, nil, reader)
}

// StreamHttpResponseWithCookies creates a 200 success reader response with custom cookies
func StreamHttpResponseWithCookies(statusCode int, contentType string, cookies []http.Cookie, reader io.Reader) *HttpStreamResponse {
	return StreamHttpResponseWithHeadersAndCookies(statusCode, contentType, nil, cookies, reader)
}

// StreamHttpResponseWithHeadersAndCookies creates a 200 success reader response with custom headers and cookies
func StreamHttpResponseWithHeadersAndCookies(statusCode int, contentType string, headers http.Header, cookies []http.Cookie, reader io.Reader) *HttpStreamResponse {
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
