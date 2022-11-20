package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/router/path"

	"github.com/ixtendio/gofre/response"
	"time"
	"unsafe"
)

// RequestDumper dumps the request (before processing) and the corresponding response in JSON format.
// It is especially useful in debugging problems.
func RequestDumper(logger func(val string)) Middleware {
	return func(handler handler.Handler) handler.Handler {
		return func(ctx context.Context, mc path.MatchingContext) (resp response.HttpResponse, err error) {
			startTime := time.Now().UnixMilli()
			// request logging
			requestMap := map[string]any{}
			requestMap["method"] = mc.R.Method
			requestMap["protocolVersion"] = mc.R.Proto
			requestMap["requestURI"] = mc.R.RequestURI
			requestMap["url"] = mc.R.URL.String()
			requestMap["remoteAddress"] = mc.R.RemoteAddr
			requestMap["transferEncoding"] = mc.R.TransferEncoding
			requestMap["contentLength"] = mc.R.ContentLength
			requestMap["host"] = mc.R.Host
			requestMap["isSecure"] = mc.R.TLS != nil
			requestMap["header"] = mc.R.Header
			requestMap["formFields"] = mc.R.Form
			var cookies []string
			for _, c := range mc.R.Cookies() {
				if c != nil {
					cookies = append(cookies, c.String())
				}
			}
			if len(cookies) > 0 {
				requestMap["cookies"] = cookies
			}

			resp, err = handler(ctx, mc)
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
