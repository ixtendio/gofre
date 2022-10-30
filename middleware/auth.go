package middleware

import (
	"context"
	"github.com/ixtendio/gofre/auth"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/request"
	"github.com/ixtendio/gofre/response"
)

type SecurityPrincipalSupplierFunc func(ctx context.Context, r *request.HttpRequest) (auth.SecurityPrincipal, error)

// SecurityPrincipalSupplier extracts the auth.SecurityPrincipal and propagate it to the context.Context
func SecurityPrincipalSupplier(sps SecurityPrincipalSupplierFunc) Middleware {
	return func(handler handler.Handler) handler.Handler {
		return func(ctx context.Context, req *request.HttpRequest) (resp response.HttpResponse, err error) {
			if securityPrincipal, err := sps(ctx, req); err != nil {
				return nil, err
			} else {
				if securityPrincipal == nil {
					return handler(ctx, req)
				}
				return handler(context.WithValue(ctx, auth.SecurityPrincipalCtxKey, securityPrincipal), req)
			}
		}
	}
}
