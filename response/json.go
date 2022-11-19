package response

import (
	"encoding/json"
	"fmt"
	"github.com/ixtendio/gofre/router/path"

	"net/http"
	"strings"
)

const jsonContentType = "application/json"

var emptyJson = []byte("{}")

type HttpJsonResponse struct {
	HttpHeadersResponse
	Payload any
}

func (r *HttpJsonResponse) Write(w http.ResponseWriter, req path.MatchingContext) error {
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
	return JsonHttpResponseWithHeadersAndCookies(http.StatusOK, payload, nil, nil)
}

// JsonHttpResponse creates a JSON response with a specific status code
func JsonHttpResponse(statusCode int, payload any) *HttpJsonResponse {
	return JsonHttpResponseWithHeadersAndCookies(statusCode, payload, nil, nil)
}

// JsonHttpResponseWithCookies creates a JSON response with a specific status code and cookies
func JsonHttpResponseWithCookies(statusCode int, payload any, cookies HttpCookies) *HttpJsonResponse {
	return JsonHttpResponseWithHeadersAndCookies(statusCode, payload, nil, cookies)
}

// JsonHttpResponseWithHeaders creates a JSON response with a specific status code and headers
func JsonHttpResponseWithHeaders(statusCode int, payload any, headers HttpHeaders) *HttpJsonResponse {
	return JsonHttpResponseWithHeadersAndCookies(statusCode, payload, headers, nil)
}

// JsonHttpResponseWithHeadersAndCookies creates a JSON response with a specific status code, custom headers and cookies
func JsonHttpResponseWithHeadersAndCookies(statusCode int, payload any, headers HttpHeaders, cookies HttpCookies) *HttpJsonResponse {
	return &HttpJsonResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			ContentType:    jsonContentType,
			HttpHeaders:    headers,
			HttpCookies:    cookies,
		},
		Payload: payload,
	}
}

// JsonErrorHttpResponse creates an error JSON response
func JsonErrorHttpResponse(statusCode int, err error) *HttpJsonResponse {
	return JsonErrorHttpResponseWithHeadersAndCookies(statusCode, err, nil, nil)
}

// JsonErrorHttpResponseWithCookies creates an error JSON response with custom cookies
// The cookies, if present, once will be written to output will be added in the pool for re-use
func JsonErrorHttpResponseWithCookies(statusCode int, err error, cookies HttpCookies) *HttpJsonResponse {
	return JsonErrorHttpResponseWithHeadersAndCookies(statusCode, err, nil, cookies)
}

// JsonErrorHttpResponseWithHeaders creates an error JSON response with custom headers
// The headers, if present, once will be written to output will be added in the pool for re-use
func JsonErrorHttpResponseWithHeaders(statusCode int, err error, headers HttpHeaders) *HttpJsonResponse {
	return JsonErrorHttpResponseWithHeadersAndCookies(statusCode, err, headers, nil)
}

// JsonErrorHttpResponseWithHeadersAndCookies creates an error JSON response with custom headers and cookies
// The headers and cookies, if present, once will be written to output will be added in the pool for re-use
func JsonErrorHttpResponseWithHeadersAndCookies(statusCode int, err error, headers HttpHeaders, cookies HttpCookies) *HttpJsonResponse {
	return &HttpJsonResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			ContentType:    jsonContentType,
			HttpHeaders:    headers,
			HttpCookies:    cookies,
		},
		Payload: map[string]string{"error": errorToString(err)},
	}
}

func errorToString(err error) string {
	if err == nil || len(err.Error()) == 0 {
		return ""
	}
	errMsg := err.Error()
	return strings.ToUpper(errMsg[0:1]) + errMsg[1:]
}
