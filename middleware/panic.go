package middleware

import (
	"context"
	"fmt"
	"runtime/debug"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/request"
	"github.com/ixtendio/gofre/response"
)

// PanicRecover is middleware that recovers from panic and convert it to an error
func PanicRecover() Middleware {
	return func(handler handler.Handler) handler.Handler {
		return func(ctx context.Context, req *request.HttpRequest) (resp response.HttpResponse, err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("recover from panic, err: %w, stack: [%s]", r, string(debug.Stack()))
				}
			}()
			resp, err = handler(ctx, req)
			return resp, err
		}
	}
}
