package response

import (
	"encoding/json"
	"fmt"
	"github.com/ixtendio/gow/request"
	"net/http"
	"unicode"
)

var defaultJsonHeaders = map[string]string{
	"Content-Type": "application/json",
}
var emptyJson = []byte("{}")

type HttpJsonResponse struct {
	HttpHeadersResponse
	Payload any
}

func (r *HttpJsonResponse) Write(w http.ResponseWriter, responseContext *request.HttpRequest) error {
	// write the headers
	if err := r.HttpHeadersResponse.Write(w, responseContext); err != nil {
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
			HttpHeaders:    defaultJsonHeaders,
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
			HttpHeaders:    defaultJsonHeaders,
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
			HttpHeaders:    defaultJsonHeaders,
			HttpCookies:    cookies,
		},
		Payload: payload,
	}
}

// JsonHttpResponseWithHeaders creates a JSON response with a specific status code and headers
func JsonHttpResponseWithHeaders(statusCode int, payload any, headers map[string]string) *HttpJsonResponse {
	headers["Content-Type"] = "application/json"
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
func JsonHttpResponseWithHeadersAndCookies(statusCode int, payload any, headers map[string]string, cookies []*http.Cookie) *HttpJsonResponse {
	headers["Content-Type"] = "application/json"
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
			HttpHeaders:    defaultJsonHeaders,
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
			HttpHeaders:    defaultJsonHeaders,
			HttpCookies:    cookies,
		},
		Payload: map[string]string{"error": errorToString(err)},
	}
}

// JsonErrorHttpResponseWithHeaders creates an error JSON response with custom headers
func JsonErrorHttpResponseWithHeaders(statusCode int, err error, headers map[string]string) *HttpJsonResponse {
	headers["Content-Type"] = "application/json"
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
func JsonErrorHttpResponseWithHeadersAndCookies(statusCode int, err error, headers map[string]string, cookies []*http.Cookie) *HttpJsonResponse {
	headers["Content-Type"] = "application/json"
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
