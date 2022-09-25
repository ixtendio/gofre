package response

import (
	"fmt"
	"github.com/ixtendio/gow/request"
	"net/http"
)

var defaultPlainTextHeaders = map[string]string{
	"Content-Type": "text/plain; charset=utf-8",
}

type HttpTextResponse struct {
	HttpHeadersResponse
	Payload string
}

func (r *HttpTextResponse) Write(w http.ResponseWriter, responseContext *request.HttpRequest) error {
	// write the headers
	if err := r.HttpHeadersResponse.Write(w, responseContext); err != nil {
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
	return &HttpTextResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: http.StatusOK,
			HttpHeaders:    defaultPlainTextHeaders,
			HttpCookies:    nil,
		},
		Payload: payload,
	}
}

// HtmlHttpResponseOK creates a 200 success HTML response
func HtmlHttpResponseOK(payload string) *HttpTextResponse {
	return &HttpTextResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: http.StatusOK,
			HttpHeaders:    defaultHtmlHeaders,
			HttpCookies:    nil,
		},
		Payload: payload,
	}
}

// PlainTextHttpResponse creates a plain text response with a specific status code
func PlainTextHttpResponse(statusCode int, payload string) *HttpTextResponse {
	return &HttpTextResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    defaultPlainTextHeaders,
			HttpCookies:    nil,
		},
		Payload: payload,
	}
}

// HtmlHttpResponse creates an HTML response with a specific status code
func HtmlHttpResponse(statusCode int, payload string) *HttpTextResponse {
	return &HttpTextResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    defaultHtmlHeaders,
			HttpCookies:    nil,
		},
		Payload: payload,
	}
}

// PlainTextHttpResponseWithHeaders creates a plain text response with a specific status code and headers
func PlainTextHttpResponseWithHeaders(statusCode int, payload string, headers map[string]string) *HttpTextResponse {
	headers["Content-Type"] = defaultPlainTextHeaders["Content-Type"]
	return &HttpTextResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    headers,
			HttpCookies:    nil,
		},
		Payload: payload,
	}
}

// HtmlHttpResponseWithHeaders creates an HTML response with a specific status code and headers
func HtmlHttpResponseWithHeaders(statusCode int, payload string, headers map[string]string) *HttpTextResponse {
	headers["Content-Type"] = defaultHtmlHeaders["Content-Type"]
	return &HttpTextResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    headers,
			HttpCookies:    nil,
		},
		Payload: payload,
	}
}

// PlainTextResponseWithHeadersAndCookies creates a plain text response with a specific status code, custom headers and cookies
func PlainTextResponseWithHeadersAndCookies(statusCode int, payload string, headers map[string]string, cookies []*http.Cookie) *HttpTextResponse {
	headers["Content-Type"] = defaultPlainTextHeaders["Content-Type"]
	return &HttpTextResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    headers,
			HttpCookies:    cookies,
		},
		Payload: payload,
	}
}

// HtmlResponseWithHeadersAndCookies creates a plain text response with a specific status code, custom headers and cookies
func HtmlResponseWithHeadersAndCookies(statusCode int, payload string, headers map[string]string, cookies []*http.Cookie) *HttpTextResponse {
	headers["Content-Type"] = defaultHtmlHeaders["Content-Type"]
	return &HttpTextResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    headers,
			HttpCookies:    cookies,
		},
		Payload: payload,
	}
}
