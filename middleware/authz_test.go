package middleware

import (
	"context"
	"github.com/ixtendio/gofre/auth"
	"github.com/ixtendio/gofre/errors"
	"github.com/ixtendio/gofre/request"
	"github.com/ixtendio/gofre/response"
	"testing"
)

func TestAuthorizeAll(t *testing.T) {
	type args struct {
		ctx         context.Context
		permissions []auth.Permission
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "user not authenticated",
			args: args{
				ctx:         context.Background(),
				permissions: []auth.Permission{{Scope: "domain/subdomain/resource", Access: auth.AccessRead}},
			},
			want: errors.ErrUnauthorizedRequest,
		},
		{
			name: "user has all permissions",
			args: args{
				ctx: context.WithValue(context.Background(), auth.SecurityPrincipalCtxKey, auth.User{
					Groups: []auth.Group{{
						Roles: []auth.Role{{
							AllowedPermissions: []auth.Permission{{
								Scope:  "domain/subdomain/*",
								Access: auth.AccessRead,
							}},
						}},
					}},
				}),
				permissions: []auth.Permission{
					{Scope: "domain/subdomain/resource1", Access: auth.AccessRead},
					{Scope: "domain/subdomain/resource2", Access: auth.AccessRead}},
			},
			want: nil,
		},
		{
			name: "user has not all permissions",
			args: args{
				ctx: context.WithValue(context.Background(), auth.SecurityPrincipalCtxKey, auth.User{
					Groups: []auth.Group{{
						Roles: []auth.Role{{
							AllowedPermissions: []auth.Permission{{
								Scope:  "domain/subdomain/resource1",
								Access: auth.AccessRead,
							}},
						}},
					}},
				}),
				permissions: []auth.Permission{
					{Scope: "domain/subdomain/resource1", Access: auth.AccessRead},
					{Scope: "domain/subdomain/resource2", Access: auth.AccessRead}},
			},
			want: errors.ErrUnauthorizedRequest,
		},
		{
			name: "user has no permissions",
			args: args{
				ctx: context.WithValue(context.Background(), auth.SecurityPrincipalCtxKey, auth.User{}),
				permissions: []auth.Permission{
					{Scope: "domain/subdomain/resource1", Access: auth.AccessRead},
					{Scope: "domain/subdomain/resource2", Access: auth.AccessRead}},
			},
			want: errors.ErrUnauthorizedRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := AuthorizeAll(tt.args.permissions...)(func(ctx context.Context, r request.HttpRequest) (response.HttpResponse, error) {
				return response.PlainTextHttpResponseOK("ok"), nil
			})(tt.args.ctx, request.HttpRequest{})
			if err != tt.want {
				t.Errorf("AuthorizeAll() = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestAuthorizeAny(t *testing.T) {
	type args struct {
		ctx         context.Context
		permissions []auth.Permission
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "user not authenticated",
			args: args{
				ctx:         context.Background(),
				permissions: []auth.Permission{{Scope: "domain/subdomain/resource", Access: auth.AccessRead}},
			},
			want: errors.ErrUnauthorizedRequest,
		},
		{
			name: "user has all permissions",
			args: args{
				ctx: context.WithValue(context.Background(), auth.SecurityPrincipalCtxKey, auth.User{
					Groups: []auth.Group{{
						Roles: []auth.Role{{
							AllowedPermissions: []auth.Permission{{
								Scope:  "domain/subdomain/*",
								Access: auth.AccessRead,
							}},
						}},
					}},
				}),
				permissions: []auth.Permission{
					{Scope: "domain/subdomain/resource1", Access: auth.AccessRead},
					{Scope: "domain/subdomain/resource2", Access: auth.AccessRead}},
			},
			want: nil,
		},
		{
			name: "user has only a subset of permissions",
			args: args{
				ctx: context.WithValue(context.Background(), auth.SecurityPrincipalCtxKey, auth.User{
					Groups: []auth.Group{{
						Roles: []auth.Role{{
							AllowedPermissions: []auth.Permission{{
								Scope:  "domain/subdomain/resource1",
								Access: auth.AccessRead,
							}},
						}},
					}},
				}),
				permissions: []auth.Permission{
					{Scope: "domain/subdomain/resource1", Access: auth.AccessRead},
					{Scope: "domain/subdomain/resource2", Access: auth.AccessRead}},
			},
			want: nil,
		},
		{
			name: "user has no permissions",
			args: args{
				ctx: context.WithValue(context.Background(), auth.SecurityPrincipalCtxKey, auth.User{}),
				permissions: []auth.Permission{
					{Scope: "domain/subdomain/resource1", Access: auth.AccessRead},
					{Scope: "domain/subdomain/resource2", Access: auth.AccessRead}},
			},
			want: errors.ErrUnauthorizedRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := AuthorizeAny(tt.args.permissions...)(func(ctx context.Context, r request.HttpRequest) (response.HttpResponse, error) {
				return response.PlainTextHttpResponseOK("ok"), nil
			})(tt.args.ctx, request.HttpRequest{})
			if err != tt.want {
				t.Errorf("AuthorizeAll() = %v, want %v", err, tt.want)
			}
		})
	}
}
