package middleware

import (
	"context"
	"fmt"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/router/path"

	"github.com/ixtendio/gofre/response"
	"runtime/debug"
)

// PanicRecover is middleware that recovers from panic and convert it to an error
func PanicRecover() Middleware {
	return func(handler handler.Handler) handler.Handler {
		return func(ctx context.Context, mc path.MatchingContext) (resp response.HttpResponse, err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("recover from panic, stack: [%s]", string(debug.Stack()))
				}
			}()
			resp, err = handler(ctx, mc)
			return resp, err
		}
	}
}
