package response

import (
	"fmt"
	"github.com/ixtendio/gow/request"
	"io"
	"net/http"
)

type HttpStreamResponse struct {
	HttpHeadersResponse
	Reader io.Reader
}

func (r *HttpStreamResponse) Write(w http.ResponseWriter, reqContext *request.HttpRequest) error {
	// write the headers
	if err := r.HttpHeadersResponse.Write(w, reqContext); err != nil {
		return err
	}

	if _, err := io.Copy(w, r.Reader); err != nil {
		return fmt.Errorf("failed transferig the input stream, err: %w", err)
	}
	return nil
}

// StreamHttpResponse creates a 200 success reader response with a specific content type
func StreamHttpResponse(reader io.Reader, contentType string) *HttpStreamResponse {
	return &HttpStreamResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: http.StatusOK,
			HttpHeaders: map[string]string{
				"Content-Type": contentType,
			},
			HttpCookies: nil,
		},
		Reader: reader,
	}
}

// StreamHttpResponseWithHeaders creates a 200 success reader response with custom headers
func StreamHttpResponseWithHeaders(statusCode int, reader io.Reader, contentType string, headers map[string]string) *HttpStreamResponse {
	headers["Content-Type"] = contentType
	return &HttpStreamResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    headers,
			HttpCookies:    nil,
		},
		Reader: reader,
	}
}

// StreamHttpResponseWithCookies creates a 200 success reader response with custom cookies
func StreamHttpResponseWithCookies(statusCode int, reader io.Reader, contentType string, cookies []*http.Cookie) *HttpStreamResponse {
	return &HttpStreamResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders: map[string]string{
				"Content-Type": contentType,
			},
			HttpCookies: cookies,
		},
		Reader: reader,
	}
}

// StreamHttpResponseWithHeadersAndCookies creates a 200 success reader response with custom headers and cookies
func StreamHttpResponseWithHeadersAndCookies(statusCode int, reader io.Reader, contentType string, headers map[string]string, cookies []*http.Cookie) *HttpStreamResponse {
	headers["Content-Type"] = contentType
	return &HttpStreamResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    headers,
			HttpCookies:    cookies,
		},
		Reader: reader,
	}
}
