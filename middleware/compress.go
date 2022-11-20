package middleware

import (
	"context"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/router/path"

	"github.com/ixtendio/gofre/response"
)

// CompressResponse enables the HTTP response compressing as long as the client support it via `Accept-Encoding` request header
func CompressResponse(compressionLevel int) Middleware {
	return func(handler handler.Handler) handler.Handler {
		return func(ctx context.Context, mc path.MatchingContext) (resp response.HttpResponse, err error) {
			resp, err = handler(ctx, mc)
			if err != nil {
				return nil, err
			}
			return response.NewHttpCompressResponse(resp, compressionLevel)
		}
	}
}
