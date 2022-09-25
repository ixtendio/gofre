package middleware

import (
	"context"
	"fmt"
	"github.com/ixtendio/gow/request"
	"github.com/ixtendio/gow/response"
	"github.com/ixtendio/gow/router"
)

func Panic() Middleware {
	return func(handler router.Handler) router.Handler {
		return func(ctx context.Context, req *request.HttpRequest) (resp response.HttpResponse, err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("recover from panic, err: %v", r)
				}
			}()
			resp, err = handler(ctx, req)
			return resp, err
		}
	}
}
