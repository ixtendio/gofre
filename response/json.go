package response

import (
	"encoding/json"
	"fmt"
	"github.com/ixtendio/gofre/request"
	"net/http"
	"unicode"
)

const jsonContentType = "application/json"

var defaultJsonHeaders = func() http.Header {
	return http.Header{"Content-Type": {jsonContentType}}
}

var emptyJson = []byte("{}")

type HttpJsonResponse struct {
	HttpHeadersResponse
	Payload any
}

func (r *HttpJsonResponse) Write(w http.ResponseWriter, req *request.HttpRequest) error {
	// write the headers
	if err := r.HttpHeadersResponse.Write(w, req); err != nil {
		return err
	}

	var err error
	var payload []byte
	if r.Payload != nil {
		payload, err = json.Marshal(r.Payload)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON response, err: %w", err)
		}
	} else {
		payload = emptyJson
	}

	// write the JSON response
	if _, err := w.Write(payload); err != nil {
		return fmt.Errorf("failed to write the JSON response, err: %w", err)
	}
	return nil
}

// JsonHttpResponseOK creates a 200 success JSON response
func JsonHttpResponseOK(payload any) *HttpJsonResponse {
	return &HttpJsonResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: http.StatusOK,
			HttpHeaders:    defaultJsonHeaders(),
			HttpCookies:    nil,
		},
		Payload: payload,
	}
}

// JsonHttpResponse creates a JSON response with a specific status code
func JsonHttpResponse(statusCode int, payload any) *HttpJsonResponse {
	return &HttpJsonResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    defaultJsonHeaders(),
			HttpCookies:    nil,
		},
		Payload: payload,
	}
}

// JsonHttpResponseWithCookies creates a JSON response with a specific status code and cookies
func JsonHttpResponseWithCookies(statusCode int, payload any, cookies []*http.Cookie) *HttpJsonResponse {
	return &HttpJsonResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    defaultJsonHeaders(),
			HttpCookies:    cookies,
		},
		Payload: payload,
	}
}

// JsonHttpResponseWithHeaders creates a JSON response with a specific status code and headers
func JsonHttpResponseWithHeaders(statusCode int, payload any, headers http.Header) *HttpJsonResponse {
	headers.Set("Content-Type", jsonContentType)
	return &HttpJsonResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    headers,
			HttpCookies:    nil,
		},
		Payload: payload,
	}
}

// JsonHttpResponseWithHeadersAndCookies creates a JSON response with a specific status code, custom headers and cookies
func JsonHttpResponseWithHeadersAndCookies(statusCode int, payload any, headers http.Header, cookies []*http.Cookie) *HttpJsonResponse {
	headers.Set("Content-Type", jsonContentType)
	return &HttpJsonResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    headers,
			HttpCookies:    cookies,
		},
		Payload: payload,
	}
}

// JsonErrorHttpResponse creates an error JSON response
func JsonErrorHttpResponse(statusCode int, err error) *HttpJsonResponse {
	return &HttpJsonResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    defaultJsonHeaders(),
			HttpCookies:    nil,
		},
		Payload: map[string]string{"error": errorToString(err)},
	}
}

// JsonErrorHttpResponseWithCookies creates an error JSON response with custom cookies
func JsonErrorHttpResponseWithCookies(statusCode int, err error, cookies []*http.Cookie) *HttpJsonResponse {
	return &HttpJsonResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    defaultJsonHeaders(),
			HttpCookies:    cookies,
		},
		Payload: map[string]string{"error": errorToString(err)},
	}
}

// JsonErrorHttpResponseWithHeaders creates an error JSON response with custom headers
func JsonErrorHttpResponseWithHeaders(statusCode int, err error, headers http.Header) *HttpJsonResponse {
	headers.Set("Content-Type", jsonContentType)
	return &HttpJsonResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    headers,
			HttpCookies:    nil,
		},
		Payload: map[string]string{"error": errorToString(err)},
	}
}

// JsonErrorHttpResponseWithHeadersAndCookies creates an error JSON response with custom headers and cookies
func JsonErrorHttpResponseWithHeadersAndCookies(statusCode int, err error, headers http.Header, cookies []*http.Cookie) *HttpJsonResponse {
	headers.Set("Content-Type", jsonContentType)
	return &HttpJsonResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    headers,
			HttpCookies:    cookies,
		},
		Payload: map[string]string{"error": errorToString(err)},
	}
}

func errorToString(err error) string {
	errorMsgRune := []rune(err.Error())
	errorMsgRune[0] = unicode.ToUpper(errorMsgRune[0])
	return string(errorMsgRune)
}
