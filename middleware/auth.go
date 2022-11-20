package middleware

import (
	"context"
	"github.com/ixtendio/gofre/auth"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/router/path"

	"github.com/ixtendio/gofre/response"
)

type SecurityPrincipalSupplierFunc func(ctx context.Context, mc path.MatchingContext) (auth.SecurityPrincipal, error)

// SecurityPrincipalSupplier extracts the auth.SecurityPrincipal and propagate it to the context.Context
func SecurityPrincipalSupplier(sps SecurityPrincipalSupplierFunc) Middleware {
	return func(handler handler.Handler) handler.Handler {
		return func(ctx context.Context, mc path.MatchingContext) (resp response.HttpResponse, err error) {
			if securityPrincipal, err := sps(ctx, mc); err != nil {
				return nil, err
			} else {
				if securityPrincipal == nil {
					return handler(ctx, mc)
				}
				return handler(context.WithValue(ctx, auth.SecurityPrincipalCtxKey, securityPrincipal), mc)
			}
		}
	}
}
