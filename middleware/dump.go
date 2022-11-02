package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/request"
	"github.com/ixtendio/gofre/response"
	"time"
	"unsafe"
)

// RequestDumper dumps the request (before processing) and the corresponding response in JSON format.
// It is especially useful in debugging problems.
func RequestDumper(logger func(val string)) Middleware {
	return func(handler handler.Handler) handler.Handler {
		return func(ctx context.Context, req *request.HttpRequest) (resp response.HttpResponse, err error) {
			startTime := time.Now().UnixMilli()
			// request logging
			requestMap := map[string]any{}
			requestMap["method"] = req.R.Method
			requestMap["protocolVersion"] = req.R.Proto
			requestMap["requestURI"] = req.R.RequestURI
			requestMap["url"] = req.R.URL.String()
			requestMap["remoteAddress"] = req.R.RemoteAddr
			requestMap["transferEncoding"] = req.R.TransferEncoding
			requestMap["contentLength"] = req.R.ContentLength
			requestMap["host"] = req.R.Host
			requestMap["isSecure"] = req.R.TLS != nil
			requestMap["header"] = req.R.Header
			requestMap["formFields"] = req.R.Form
			var cookies []string
			for _, c := range req.R.Cookies() {
				if c != nil {
					cookies = append(cookies, c.String())
				}
			}
			if len(cookies) > 0 {
				requestMap["cookies"] = cookies
			}

			resp, err = handler(ctx, req)
			if err != nil {
				return nil, err
			}

			// response logging
			responseMap := map[string]any{}
			responseMap["statusCode"] = resp.StatusCode()
			responseMap["header"] = resp.Headers()
			cookies = nil
			for _, c := range resp.Cookies() {
				cookies = append(cookies, c.String())
			}
			if len(cookies) > 0 {
				responseMap["cookies"] = cookies
			}

			data, err := json.Marshal(map[string]any{
				"startTimeMs": startTime,
				"endTimeMs":   time.Now().UnixMilli(),
				"request":     requestMap,
				"response":    responseMap,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to dump the request, err: %w", err)
			}
			logger(*(*string)(unsafe.Pointer(&data)))
			return resp, err
		}
	}
}
