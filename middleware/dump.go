package middleware

import (
	"context"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/request"
	"github.com/ixtendio/gofre/response"
	"strings"
)

// RequestDumper dumps the request (before processing) and the corresponding response.
// It is especially useful in debugging problems.
func RequestDumper(logResponsePayload bool, logger func(val string)) Middleware {
	return func(handler handler.Handler) handler.Handler {
		return func(ctx context.Context, req *request.HttpRequest) (resp response.HttpResponse, err error) {
			var sb strings.Builder
			resp, err = handler(ctx, req)
			logger(sb.String())
			return resp, err
		}
	}
}
