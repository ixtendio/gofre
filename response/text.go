package response

import (
	"fmt"
	"github.com/ixtendio/gofre/request"
	"net/http"
)

type HttpTextResponse struct {
	HttpHeadersResponse
	Payload string
}

func (r *HttpTextResponse) Write(w http.ResponseWriter, req request.HttpRequest) error {
	// write the headers
	if err := r.HttpHeadersResponse.Write(w, req); err != nil {
		return err
	}

	// write the JSON response
	if _, err := w.Write([]byte(r.Payload)); err != nil {
		return fmt.Errorf("failed to write the text response, err: %w", err)
	}
	return nil
}

// PlainTextHttpResponseOK creates a 200 success plain text response
func PlainTextHttpResponseOK(payload string) *HttpTextResponse {
	return PlainTextResponseWithHeadersAndCookies(http.StatusOK, payload, nil, nil)
}

// PlainTextHttpResponse creates a plain text response with a specific status code
func PlainTextHttpResponse(statusCode int, payload string) *HttpTextResponse {
	return PlainTextResponseWithHeadersAndCookies(statusCode, payload, nil, nil)
}

// PlainTextHttpResponseWithHeaders creates a plain text response with a specific status code and headers
func PlainTextHttpResponseWithHeaders(statusCode int, payload string, headers http.Header) *HttpTextResponse {
	return PlainTextResponseWithHeadersAndCookies(statusCode, payload, headers, nil)
}

// PlainTextResponseWithHeadersAndCookies creates a plain text response with a specific status code, custom headers and cookies
func PlainTextResponseWithHeadersAndCookies(statusCode int, payload string, headers http.Header, cookies []http.Cookie) *HttpTextResponse {
	if headers == nil {
		headers = http.Header{}
	}
	headers.Set("Content-Type", plainTextContentType)
	return &HttpTextResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    headers,
			HttpCookies:    NewHttpCookies(cookies),
		},
		Payload: payload,
	}
}

// HtmlHttpResponseOK creates a 200 success HTML response
func HtmlHttpResponseOK(payload string) *HttpTextResponse {
	return HtmlResponseWithHeadersAndCookies(http.StatusOK, payload, nil, nil)
}

// HtmlHttpResponse creates an HTML response with a specific status code
func HtmlHttpResponse(statusCode int, payload string) *HttpTextResponse {
	return HtmlResponseWithHeadersAndCookies(statusCode, payload, nil, nil)
}

// HtmlHttpResponseWithHeaders creates an HTML response with a specific status code and headers
func HtmlHttpResponseWithHeaders(statusCode int, payload string, headers http.Header) *HttpTextResponse {
	return HtmlResponseWithHeadersAndCookies(statusCode, payload, headers, nil)
}

// HtmlResponseWithHeadersAndCookies creates a plain text response with a specific status code, custom headers and cookies
func HtmlResponseWithHeadersAndCookies(statusCode int, payload string, headers http.Header, cookies []http.Cookie) *HttpTextResponse {
	if headers == nil {
		headers = http.Header{}
	}
	headers.Set("Content-Type", htmlContentType)
	return &HttpTextResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    headers,
			HttpCookies:    NewHttpCookies(cookies),
		},
		Payload: payload,
	}
}
