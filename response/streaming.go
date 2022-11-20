package response

import (
	"fmt"
	"github.com/ixtendio/gofre/router/path"

	"io"
	"net/http"
)

type HttpStreamResponse struct {
	HttpHeadersResponse
	Reader io.Reader
}

func (r *HttpStreamResponse) Write(w http.ResponseWriter, mc path.MatchingContext) error {
	// write the headers
	if err := r.HttpHeadersResponse.Write(w, mc); err != nil {
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
// The headers, if present, once will be written to output will be added in the pool for re-use
func StreamHttpResponseWithHeaders(statusCode int, contentType string, headers HttpHeaders, reader io.Reader) *HttpStreamResponse {
	return StreamHttpResponseWithHeadersAndCookies(statusCode, contentType, headers, nil, reader)
}

// StreamHttpResponseWithCookies creates a 200 success reader response with custom cookies
// The cookies, if present, once will be written to output will be added in the pool for re-use
func StreamHttpResponseWithCookies(statusCode int, contentType string, cookies HttpCookies, reader io.Reader) *HttpStreamResponse {
	return StreamHttpResponseWithHeadersAndCookies(statusCode, contentType, nil, cookies, reader)
}

// StreamHttpResponseWithHeadersAndCookies creates a 200 success reader response with custom headers and cookies
// The headers and cookies, if present, once will be written to output will be added in the pool for re-use
func StreamHttpResponseWithHeadersAndCookies(statusCode int, contentType string, headers HttpHeaders, cookies HttpCookies, reader io.Reader) *HttpStreamResponse {
	return &HttpStreamResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    headers,
			ContentType:    contentType,
			HttpCookies:    cookies,
		},
		Reader: reader,
	}
}
