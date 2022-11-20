package middleware

import (
	"context"
	"errors"
	"github.com/ixtendio/gofre/auth"
	"github.com/ixtendio/gofre/router/path"

	"github.com/ixtendio/gofre/response"
	"net/http"
	"reflect"
	"testing"
)

func TestAuthProvider(t *testing.T) {
	oldUsr := auth.User{
		Id:               "user1",
		Name:             "",
		IdentityPlatform: "",
		Groups:           nil,
	}
	usr := auth.User{
		Id:               "user2",
		Name:             "Gorex",
		IdentityPlatform: "GCP",
		Groups:           nil,
	}
	type args struct {
		ctx context.Context
		sps SecurityPrincipalSupplierFunc
	}
	tests := []struct {
		name    string
		args    args
		want    auth.SecurityPrincipal
		wantErr bool
	}{
		{
			name: "SecurityPrincipalSupplierFunc returns ok and overrides the existing one",
			args: args{
				ctx: context.WithValue(context.Background(), auth.SecurityPrincipalCtxKey, oldUsr),
				sps: func(ctx context.Context, mc path.MatchingContext) (auth.SecurityPrincipal, error) {
					return usr, nil
				},
			},
			want:    usr,
			wantErr: false,
		},
		{
			name: "SecurityPrincipalSupplierFunc returns ok",
			args: args{
				ctx: context.Background(),
				sps: func(ctx context.Context, mc path.MatchingContext) (auth.SecurityPrincipal, error) {
					return usr, nil
				},
			},
			want:    usr,
			wantErr: false,
		},
		{
			name: "SecurityPrincipalSupplierFunc returns error",
			args: args{
				ctx: context.Background(),
				sps: func(ctx context.Context, mc path.MatchingContext) (auth.SecurityPrincipal, error) {
					return nil, errors.New("an error")
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "SecurityPrincipalSupplierFunc returns nil",
			args: args{
				ctx: context.Background(),
				sps: func(ctx context.Context, mc path.MatchingContext) (auth.SecurityPrincipal, error) {
					return nil, nil
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequestWithContext(tt.args.ctx, "GET", "https://www.domain.com", nil)
			var gotSecurityPrincipal auth.SecurityPrincipal
			_, err := SecurityPrincipalSupplier(tt.args.sps)(func(ctx context.Context, mc path.MatchingContext) (response.HttpResponse, error) {
				gotSecurityPrincipal = auth.GetSecurityPrincipalFromContext(ctx)
				return response.PlainTextHttpResponseOK("ok"), nil
			})(tt.args.ctx, path.MatchingContext{R: req})
			if tt.wantErr {
				if err == nil {
					t.Errorf("SecurityPrincipalSupplier() want error got nil")
				}
			} else {
				if !reflect.DeepEqual(gotSecurityPrincipal, tt.want) {
					t.Errorf("SecurityPrincipalSupplier() got: %v, want: %v", gotSecurityPrincipal, tt.want)
				}
			}
		})
	}
}
