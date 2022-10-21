package middleware

import (
	"context"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/request"
	"github.com/ixtendio/gofre/response"
)

func CompressResponse(compressionLevel int) Middleware {
	return func(handler handler.Handler) handler.Handler {
		return func(ctx context.Context, req *request.HttpRequest) (resp response.HttpResponse, err error) {
			resp, err = handler(ctx, req)
			if err != nil {
				return nil, err
			}
			return response.NewHttpCompressResponse(resp, compressionLevel)
		}
	}
}
