package middleware

import (
	"context"
	"github.com/ixtendio/gow/auth"
	"github.com/ixtendio/gow/errors"
	"github.com/ixtendio/gow/request"
	"github.com/ixtendio/gow/response"
	"github.com/ixtendio/gow/router"
)

// AuthorizeAll checks if the authenticated auth.SecurityPrincipal has all the requested permissions
// An errors.ErrUnauthorized error is returned if not all permissions are allowed to be executed
// by the current auth.SecurityPrincipal
func AuthorizeAll(permissions ...auth.Permission) Middleware {
	return func(handler router.Handler) router.Handler {
		return func(ctx context.Context, req *request.HttpRequest) (resp response.HttpResponse, err error) {
			securityPrincipal := auth.GetSecurityPrincipalFromContext(ctx)
			if securityPrincipal == nil {
				return nil, errors.ErrUnauthorized
			}
			for _, permission := range permissions {
				if !securityPrincipal.HasPermission(permission) {
					return nil, errors.ErrUnauthorized
				}
			}
			return handler(ctx, req)
		}
	}
}

// AuthorizeAny checks if the authenticated auth.SecurityPrincipal has at least one from the requested permissions
// An errors.ErrUnauthorized error is returned if not at least one permission is allowed to be executed
// by the current auth.SecurityPrincipal
func AuthorizeAny(permissions ...auth.Permission) Middleware {
	return func(handler router.Handler) router.Handler {
		return func(ctx context.Context, req *request.HttpRequest) (resp response.HttpResponse, err error) {
			securityPrincipal := auth.GetSecurityPrincipalFromContext(ctx)
			if securityPrincipal == nil {
				return nil, errors.ErrUnauthorized
			}
			for _, permission := range permissions {
				if securityPrincipal.HasPermission(permission) {
					return handler(ctx, req)
				}
			}
			return nil, errors.ErrUnauthorized
		}
	}
}
