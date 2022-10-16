package middleware

import (
	"context"
	"github.com/ixtendio/gofre/request"
	"github.com/ixtendio/gofre/response"
	"net/http"
	"reflect"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	type args struct {
		config SecurityHeadersConfig
		req    *http.Request
	}
	tests := []struct {
		name                string
		args                args
		wantResponseHeaders http.Header
	}{
		{
			name:                "empty config",
			args:                args{config: SecurityHeadersConfig{}, req: &http.Request{URL: mustParseURL("https://domain.com")}},
			wantResponseHeaders: http.Header{"Content-Type": {"text/plain; charset=utf-8"}},
		},
		{
			name: "HSTS config",
			args: args{config: SecurityHeadersConfig{
				STS: ShStrictTransportSecurityConfig{
					Enabled:           true,
					MaxAgeSeconds:     10,
					IncludeSubDomains: true,
					Preload:           true,
					headerValue:       "",
				}}, req: &http.Request{URL: mustParseURL("https://domain.com")}},
			wantResponseHeaders: http.Header{
				"Content-Type": {"text/plain; charset=utf-8"},
				stsHeaderName:  {"max-age=10;includeSubDomains;preload"},
			},
		},
		{
			name: "HSTS config with custom value",
			args: args{config: SecurityHeadersConfig{
				STS: ShStrictTransportSecurityConfig{
					Enabled:           true,
					MaxAgeSeconds:     10,
					IncludeSubDomains: true,
					Preload:           true,
					headerValue:       "stsHeaderValue1",
				}}, req: &http.Request{URL: mustParseURL("https://domain.com")}},
			wantResponseHeaders: http.Header{
				"Content-Type": {"text/plain; charset=utf-8"},
				stsHeaderName:  {"stsHeaderValue1"},
			},
		},
		{
			name: "ClickJacking config",
			args: args{config: SecurityHeadersConfig{
				ClickJacking: ShClickJackingConfig{
					Enabled:                 true,
					XFrameOption:            XFrameOptionAllowFrom,
					XFrameOptionHeaderValue: "XFrameOptionHeaderValue1",
					XFrameAllowFromUri:      "https://domain.com",
					headerValue:             "",
				},
			}, req: &http.Request{URL: mustParseURL("https://domain.com")}},
			wantResponseHeaders: http.Header{
				"Content-Type":             {"text/plain; charset=utf-8"},
				antiClickJackingHeaderName: {"XFrameOptionHeaderValue1 https://domain.com"},
			},
		},
		{
			name: "ClickJacking config with custom header",
			args: args{config: SecurityHeadersConfig{
				ClickJacking: ShClickJackingConfig{
					Enabled:                 true,
					XFrameOption:            XFrameOptionAllowFrom,
					XFrameOptionHeaderValue: "XFrameOptionHeaderValue1",
					XFrameAllowFromUri:      "https://domain.com",
					headerValue:             "ClickJackingCustomHeaderValue",
				},
			}, req: &http.Request{URL: mustParseURL("https://domain.com")}},
			wantResponseHeaders: http.Header{
				"Content-Type":             {"text/plain; charset=utf-8"},
				antiClickJackingHeaderName: {"ClickJackingCustomHeaderValue"},
			},
		},
		{
			name: "BlockContentSniffingEnabled",
			args: args{config: SecurityHeadersConfig{
				BlockContentSniffingEnabled: true,
			}, req: &http.Request{URL: mustParseURL("https://domain.com")}},
			wantResponseHeaders: http.Header{
				"Content-Type":                     {"text/plain; charset=utf-8"},
				blockContentTypeSniffingHeaderName: {blockContentTypeSniffingHeaderValue},
			},
		},
		{
			name: "XSSProtectionEnabled",
			args: args{config: SecurityHeadersConfig{
				XSSProtectionEnabled: true,
			}, req: &http.Request{URL: mustParseURL("https://domain.com")}},
			wantResponseHeaders: http.Header{
				"Content-Type":          {"text/plain; charset=utf-8"},
				xssProtectionHeaderName: {xssProtectionHeaderValue},
			},
		},
		{
			name: "All security protections enabled",
			args: args{config: SecurityHeadersConfig{
				STS: ShStrictTransportSecurityConfig{
					Enabled:           true,
					MaxAgeSeconds:     10,
					IncludeSubDomains: true,
					Preload:           true,
					headerValue:       "",
				},
				ClickJacking: ShClickJackingConfig{
					Enabled:                 true,
					XFrameOption:            XFrameOptionAllowFrom,
					XFrameOptionHeaderValue: "XFrameOptionHeaderValue1",
					XFrameAllowFromUri:      "https://domain.com",
					headerValue:             "",
				},
				BlockContentSniffingEnabled: true,
				XSSProtectionEnabled:        true,
			}, req: &http.Request{URL: mustParseURL("https://domain.com")}},
			wantResponseHeaders: http.Header{
				"Content-Type":                     {"text/plain; charset=utf-8"},
				stsHeaderName:                      {"max-age=10;includeSubDomains;preload"},
				antiClickJackingHeaderName:         {"XFrameOptionHeaderValue1 https://domain.com"},
				blockContentTypeSniffingHeaderName: {blockContentTypeSniffingHeaderValue},
				xssProtectionHeaderName:            {xssProtectionHeaderValue},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := SecurityHeaders(tt.args.config)
			resp, err := m(func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
				return response.PlainTextHttpResponseOK("ok"), nil
			})(context.Background(), &request.HttpRequest{R: tt.args.req})

			if err != nil {
				t.Fatalf("SecurityHeaders() returned error: %v", err)
			}

			if !reflect.DeepEqual(resp.Headers(), tt.wantResponseHeaders) {
				t.Fatalf("SecurityHeaders() response headers = %v, want = %v", resp.Headers(), tt.wantResponseHeaders)
			}
		})
	}
}
