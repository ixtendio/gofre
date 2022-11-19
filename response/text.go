package response

import (
	"github.com/ixtendio/gofre/router/path"
	"net/http"
)

type HttpTextResponse struct {
	HttpHeadersResponse
	Payload string
}

func (r *HttpTextResponse) Write(w http.ResponseWriter, req path.MatchingContext) error {
	// write the headers
	if err := r.HttpHeadersResponse.Write(w, req); err != nil {
		return err
	}

	// write the response
	return writeTextResponse(w, r.Payload)
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
// The headers, if present, once will be written to output will be added in the pool for re-use
func PlainTextHttpResponseWithHeaders(statusCode int, payload string, headers HttpHeaders) *HttpTextResponse {
	return PlainTextResponseWithHeadersAndCookies(statusCode, payload, headers, nil)
}

// PlainTextResponseWithHeadersAndCookies creates a plain text response with a specific status code, custom headers and cookies
// The headers and cookies, if present, once will be written to output will be added in the pool for re-use
func PlainTextResponseWithHeadersAndCookies(statusCode int, payload string, headers HttpHeaders, cookies HttpCookies) *HttpTextResponse {
	return &HttpTextResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			ContentType:    plainTextContentType,
			HttpHeaders:    headers,
			HttpCookies:    cookies,
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
// The headers, if present, once will be written to output will be added in the pool for re-use
func HtmlHttpResponseWithHeaders(statusCode int, payload string, headers HttpHeaders) *HttpTextResponse {
	return HtmlResponseWithHeadersAndCookies(statusCode, payload, headers, nil)
}

// HtmlResponseWithHeadersAndCookies creates a plain text response with a specific status code, custom headers and cookies
// The headers and cookies, if present, once will be written to output will be added in the pool for re-use
func HtmlResponseWithHeadersAndCookies(statusCode int, payload string, headers HttpHeaders, cookies HttpCookies) *HttpTextResponse {
	return &HttpTextResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			ContentType:    htmlContentType,
			HttpHeaders:    headers,
			HttpCookies:    cookies,
		},
		Payload: payload,
	}
}
