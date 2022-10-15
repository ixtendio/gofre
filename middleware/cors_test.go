package middleware

import (
	"context"
	"github.com/ixtendio/gofre/errors"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/request"
	"github.com/ixtendio/gofre/response"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func Test_getMediaType(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "empty media type",
			args: "",
			want: "",
		},
		{
			name: "json media type",
			args: "application/json",
			want: "application/json",
		},
		{
			name: "html media type",
			args: "text/html; charset=utf-8",
			want: "text/html",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMediaType(tt.args); got != tt.want {
				t.Errorf("getMediaType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isSameOrigin(t *testing.T) {
	type args struct {
		reqUrl string
		origin string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "http => true",
			args: args{
				reqUrl: "http://domain.com:80",
				origin: "http://domain.com:80",
			},
			want: true,
		},
		{
			name: "http with port => true",
			args: args{
				reqUrl: "http://domain.com",
				origin: "http://domain.com:80",
			},
			want: true,
		},
		{
			name: "https with port => true",
			args: args{
				reqUrl: "https://domain.com",
				origin: "https://domain.com:443",
			},
			want: true,
		},
		{
			name: "http without scheme => false",
			args: args{
				reqUrl: "domain.com:80",
				origin: "http://domain.com:80",
			},
			want: false,
		},
		{
			name: "case insensitive => false",
			args: args{
				reqUrl: "http://domain.com:80",
				origin: "http://DOMAIN.com:80",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqUrl, err := url.Parse(tt.args.reqUrl)
			if err != nil {
				t.Errorf("Failed parsing the url: %s, err:%v", tt.args.reqUrl, err)
			}
			if got := isSameOrigin(reqUrl, tt.args.origin); got != tt.want {
				t.Errorf("isSameOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isValidOrigin(t *testing.T) {
	type args struct {
		origin string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty origin => false",
			args: args{origin: ""},
			want: false,
		},
		{
			name: "origin with % => false",
			args: args{origin: "domain%20.com"},
			want: false,
		},
		{
			name: "null origin => true",
			args: args{origin: "null"},
			want: true,
		},
		{
			name: "file:// origin => true",
			args: args{origin: "file://domain.com"},
			want: true,
		},
		{
			name: "valid origin => true",
			args: args{origin: "https://domain.com"},
			want: true,
		},
		{
			name: "invalid origin => false",
			args: args{origin: "domain.com"},
			want: false,
		},
		{
			name: "malformed url => false",
			args: args{origin: "www.domain()com"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidOrigin(tt.args.origin); got != tt.want {
				t.Errorf("isValidOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getRequestType(t *testing.T) {
	sameRawUrl := "https://ixtendio.nl"
	subdomainRawUrl := "https://subdomain.ixtendio.nl"
	URL, _ := url.Parse(sameRawUrl)

	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want corsRequestType
	}{
		{
			name: "Origin header missing => notCorsRequestType",
			args: args{r: &http.Request{Header: map[string][]string{}}},
			want: notCorsRequestType,
		},
		{
			name: "!isValidOrigin => invalidCorsRequestType",
			args: args{r: &http.Request{Header: map[string][]string{requestHeaderOrigin: {""}}}},
			want: invalidCorsRequestType,
		},
		{
			name: "isSameOrigin => notCorsRequestType",
			args: args{r: &http.Request{
				URL:    URL,
				Header: map[string][]string{requestHeaderOrigin: {sameRawUrl}}}},
			want: notCorsRequestType,
		},
		{
			name: "method GET => simpleCorsRequestType",
			args: args{r: &http.Request{
				URL:    URL,
				Method: http.MethodGet,
				Header: map[string][]string{requestHeaderOrigin: {subdomainRawUrl}}}},
			want: simpleCorsRequestType,
		},
		{
			name: "method HEAD => simpleCorsRequestType",
			args: args{r: &http.Request{
				URL:    URL,
				Method: http.MethodHead,
				Header: map[string][]string{requestHeaderOrigin: {subdomainRawUrl}}}},
			want: simpleCorsRequestType,
		},
		{
			name: "method OPTIONS => invalidCorsRequestType",
			args: args{r: &http.Request{
				URL:    URL,
				Method: http.MethodOptions,
				Header: map[string][]string{requestHeaderOrigin: {subdomainRawUrl}, requestHeaderAccessControlRequestMethod: {""}}}},
			want: invalidCorsRequestType,
		},
		{
			name: "method OPTIONS => preFlightCorsRequestType",
			args: args{r: &http.Request{
				URL:    URL,
				Method: http.MethodOptions,
				Header: map[string][]string{requestHeaderOrigin: {subdomainRawUrl}, requestHeaderAccessControlRequestMethod: {"post"}}}},
			want: preFlightCorsRequestType,
		},
		{
			name: "method OPTIONS => actualCorsRequestType",
			args: args{r: &http.Request{
				URL:    URL,
				Method: http.MethodOptions,
				Header: map[string][]string{requestHeaderOrigin: {subdomainRawUrl}}}},
			want: actualCorsRequestType,
		},
		{
			name: "method POST => simpleCorsRequestType",
			args: args{r: &http.Request{
				URL:    URL,
				Method: http.MethodPost,
				Header: map[string][]string{requestHeaderOrigin: {subdomainRawUrl}, requestHeaderContentType: {contentTypeValueTextPlain}}}},
			want: simpleCorsRequestType,
		},
		{
			name: "method POST => actualCorsRequestType",
			args: args{r: &http.Request{
				URL:    URL,
				Method: http.MethodPost,
				Header: map[string][]string{requestHeaderOrigin: {subdomainRawUrl}, requestHeaderContentType: {"image"}}}},
			want: actualCorsRequestType,
		},
		{
			name: "method POST => invalidCorsRequestType",
			args: args{r: &http.Request{
				URL:    URL,
				Method: http.MethodPost,
				Header: map[string][]string{requestHeaderOrigin: {subdomainRawUrl}, requestHeaderContentType: {""}}}},
			want: invalidCorsRequestType,
		},
		{
			name: "method DELETE => actualCorsRequestType",
			args: args{r: &http.Request{
				URL:    URL,
				Method: http.MethodDelete,
				Header: map[string][]string{requestHeaderOrigin: {subdomainRawUrl}}}},
			want: actualCorsRequestType,
		},
		{
			name: "method PUT => actualCorsRequestType",
			args: args{r: &http.Request{
				URL:    URL,
				Method: http.MethodPut,
				Header: map[string][]string{requestHeaderOrigin: {subdomainRawUrl}}}},
			want: actualCorsRequestType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRequestType(tt.args.r); got != tt.want {
				t.Errorf("getRequestType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isOriginAllowed(t *testing.T) {
	type args struct {
		originHeader string
		config       CorsConfig
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "AnyOriginAllowed is true => true",
			args: args{
				originHeader: "",
				config: CorsConfig{
					AnyOriginAllowed: true,
				},
			},
			want: true,
		},
		{
			name: "AnyOriginAllowed is false => true",
			args: args{
				originHeader: "https://www.domain.com",
				config: CorsConfig{
					AnyOriginAllowed: false,
					AllowedOrigins:   []string{"https://www.domain.com", "https://www.sub.domain.com"},
				},
			},
			want: true,
		},
		{
			name: "AnyOriginAllowed is false => false",
			args: args{
				originHeader: "https://sub.www.domain.com",
				config: CorsConfig{
					AnyOriginAllowed: false,
					AllowedOrigins:   []string{"https://www.domain.com"},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOriginAllowed(tt.args.originHeader, tt.args.config); got != tt.want {
				t.Errorf("isOriginAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_addVaryHeader(t *testing.T) {
	type args struct {
		responseHeaders http.Header
		name            string
	}
	tests := []struct {
		name string
		args args
		want http.Header
	}{
		{
			name: "add *",
			args: args{
				responseHeaders: http.Header{},
				name:            "*",
			},
			want: http.Header{"Vary": {"*"}},
		},
		{
			name: "add when header not exists",
			args: args{
				responseHeaders: http.Header{},
				name:            "val",
			},
			want: http.Header{"Vary": {"val"}},
		},
		{
			name: "add when header has no value",
			args: args{
				responseHeaders: http.Header{"Vary": {}},
				name:            "val",
			},
			want: http.Header{"Vary": {"val"}},
		},
		{
			name: "should not add when header is *",
			args: args{
				responseHeaders: http.Header{"Vary": {"*"}},
				name:            "val",
			},
			want: http.Header{"Vary": {"*"}},
		},
		{
			name: "add * when vary exists",
			args: args{
				responseHeaders: http.Header{"Vary": {"val1, val2"}},
				name:            "*",
			},
			want: http.Header{"Vary": {"*"}},
		},
		{
			name: "set single * when vary * exists",
			args: args{
				responseHeaders: http.Header{"Vary": {"val1, val2, *"}},
				name:            "val3",
			},
			want: http.Header{"Vary": {"*"}},
		},
		{
			name: "add new value when vary exists",
			args: args{
				responseHeaders: http.Header{"Vary": {"val1, val2"}},
				name:            "val3",
			},
			want: http.Header{"Vary": {"val1,val2,val3"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addVaryHeader(tt.args.responseHeaders, tt.args.name)
			if !reflect.DeepEqual(tt.want, tt.args.responseHeaders) {
				t.Errorf("addVaryHeader() = %v, want %v", tt.args.responseHeaders, tt.want)
			}
		})
	}
}

func Test_addStandardCorsHeaders(t *testing.T) {
	type args struct {
		r               *http.Request
		responseHeaders http.Header
		config          CorsConfig
	}
	tests := []struct {
		name string
		args args
		want http.Header
	}{
		{
			name: "all headers should pe present",
			args: args{
				r:               &http.Request{Method: http.MethodOptions, Header: http.Header{requestHeaderOrigin: {"orig"}}},
				responseHeaders: http.Header{},
				config: CorsConfig{
					AnyOriginAllowed:       true,
					SupportsCredentials:    true,
					ExposedHeaders:         []string{"header1", "header2"},
					PreflightMaxAgeSeconds: 10,
					AllowedHttpMethods:     []string{"post", "put"},
					AllowedHttpHeaders:     []string{"allowedHeader1", "allowedHeader2"},
				},
			},
			want: http.Header{
				responseHeaderAccessControlAllowOrigin:      {"*"},
				responseHeaderAccessControlAllowCredentials: {"true"},
				responseHeaderAccessControlExposeHeaders:    {"header1,header2"},
				responseHeaderAccessControlMaxAge:           {"10"},
				responseHeaderAccessControlAllowMethods:     {"post,put"},
				responseHeaderAccessControlAllowHeaders:     {"allowedHeader1,allowedHeader2"},
				varyHeader:                                  {requestHeaderAccessControlRequestMethod + "," + requestHeaderAccessControlRequestHeaders},
			},
		},
		{
			name: "headers for empty config",
			args: args{
				r:               &http.Request{Method: http.MethodOptions, Header: http.Header{requestHeaderOrigin: {"orig"}}},
				responseHeaders: http.Header{},
				config:          CorsConfig{},
			},
			want: http.Header{
				responseHeaderAccessControlAllowOrigin: {"orig"},
				varyHeader:                             {requestHeaderOrigin + "," + requestHeaderAccessControlRequestMethod + "," + requestHeaderAccessControlRequestHeaders},
			},
		},
		{
			name: "empty headers and method not OPTIONS",
			args: args{
				r:               &http.Request{Method: http.MethodGet, Header: http.Header{requestHeaderOrigin: {"orig"}}},
				responseHeaders: http.Header{},
				config:          CorsConfig{},
			},
			want: http.Header{
				responseHeaderAccessControlAllowOrigin: {"orig"},
				varyHeader:                             {requestHeaderOrigin},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addStandardCorsHeaders(tt.args.r, tt.args.responseHeaders, tt.args.config)
			if !reflect.DeepEqual(tt.want, tt.args.responseHeaders) {
				t.Errorf("addStandardCorsHeaders() = %v, want %v", tt.args.responseHeaders, tt.want)
			}
		})
	}
}

func Test_addPreFlightCorsHeaders(t *testing.T) {
	type args struct {
		r               *http.Request
		responseHeaders http.Header
		config          CorsConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "origin not allowed",
			args: args{
				r:               &http.Request{Header: http.Header{requestHeaderOrigin: {"orig"}}},
				responseHeaders: http.Header{},
				config:          CorsConfig{},
			},
			wantErr: errors.ErrDenied,
		},
		{
			name: "Access-Control-Request-Method header not found",
			args: args{
				r:               &http.Request{Header: http.Header{requestHeaderOrigin: {"orig"}}},
				responseHeaders: http.Header{},
				config:          CorsConfig{AllowedOrigins: []string{"orig"}},
			},
			wantErr: errors.ErrDenied,
		},
		{
			name: "Access-Control-Request-Method header not allowed",
			args: args{
				r:               &http.Request{Header: http.Header{requestHeaderOrigin: {"orig"}, requestHeaderAccessControlRequestMethod: {"post"}}},
				responseHeaders: http.Header{},
				config:          CorsConfig{AllowedOrigins: []string{"orig"}},
			},
			wantErr: errors.ErrDenied,
		},
		{
			name: "Access-Control-Request-Headers header is empty",
			args: args{
				r:               &http.Request{Header: http.Header{requestHeaderOrigin: {"orig"}, requestHeaderAccessControlRequestMethod: {"post"}}},
				responseHeaders: http.Header{},
				config:          CorsConfig{AllowedOrigins: []string{"orig"}, AllowedHttpMethods: []string{"post"}},
			},
			wantErr: nil,
		},
		{
			name: "Access-Control-Request-Headers header not allowed",
			args: args{
				r: &http.Request{Header: http.Header{
					requestHeaderOrigin:                      {"orig"},
					requestHeaderAccessControlRequestMethod:  {"post"},
					requestHeaderAccessControlRequestHeaders: {"header1"},
				}},
				responseHeaders: http.Header{},
				config: CorsConfig{
					AllowedOrigins:     []string{"orig"},
					AllowedHttpMethods: []string{"post"},
				},
			},
			wantErr: errors.ErrDenied,
		},
		{
			name: "Access-Control-Request-Headers header allowed",
			args: args{
				r: &http.Request{Header: http.Header{
					requestHeaderOrigin:                      {"orig"},
					requestHeaderAccessControlRequestMethod:  {"post"},
					requestHeaderAccessControlRequestHeaders: {"header1"},
				}},
				responseHeaders: http.Header{},
				config: CorsConfig{
					AllowedOrigins:     []string{"orig"},
					AllowedHttpMethods: []string{"post"},
					AllowedHttpHeaders: []string{"header1", "header2"},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := addPreFlightCorsHeaders(tt.args.r, tt.args.responseHeaders, tt.args.config); err != tt.wantErr {
				t.Errorf("addPreFlightCorsHeaders() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_addSimpleCorsHeaders(t *testing.T) {
	type args struct {
		r               *http.Request
		responseHeaders http.Header
		config          CorsConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "Origin not allowed",
			args: args{
				r:               &http.Request{Header: http.Header{requestHeaderOrigin: {"origin"}}},
				responseHeaders: http.Header{},
				config:          CorsConfig{},
			},
			wantErr: errors.ErrDenied,
		},
		{
			name: "HTTP method not allowed",
			args: args{
				r:               &http.Request{Method: http.MethodPut, Header: http.Header{requestHeaderOrigin: {"origin"}}},
				responseHeaders: http.Header{},
				config: CorsConfig{
					AllowedOrigins: []string{"origin"},
				},
			},
			wantErr: errors.ErrDenied,
		},
		{
			name: "origin and HTTP method allowed",
			args: args{
				r:               &http.Request{Method: http.MethodPut, Header: http.Header{requestHeaderOrigin: {"origin"}}},
				responseHeaders: http.Header{},
				config: CorsConfig{
					AllowedOrigins:     []string{"origin"},
					AllowedHttpMethods: []string{"put"},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := addSimpleCorsHeaders(tt.args.r, tt.args.responseHeaders, tt.args.config); err != tt.wantErr {
				t.Errorf("addSimpleCorsHeaders() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCors(t *testing.T) {
	type args struct {
		config CorsConfig
		req    *http.Request
	}
	tests := []struct {
		name                string
		args                args
		handler             handler.Handler
		wantError           error
		wantResponseHeaders http.Header
	}{
		{
			name: "simpleCorsRequestType success flow",
			args: args{
				config: CorsConfig{
					AllowedOrigins:     []string{"https://domain.com"},
					AllowedHttpMethods: []string{"get"},
				},
				req: &http.Request{
					Method: http.MethodGet,
					URL:    mustParseURL("https://sub.domain.com"),
					Header: http.Header{requestHeaderOrigin: {"https://domain.com"}},
				},
			},
			handler: func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
				return response.PlainTextHttpResponseOK("ok"), nil
			},
			wantError: nil,
			wantResponseHeaders: http.Header{
				responseHeaderAccessControlAllowOrigin: {"https://domain.com"},
				"Content-Type":                         {"text/plain; charset=utf-8"},
				varyHeader:                             {"Origin"}},
		},
		{
			name: "actualCorsRequestType success flow",
			args: args{
				config: CorsConfig{
					AllowedOrigins:     []string{"https://domain.com"},
					AllowedHttpMethods: []string{"options", "get"},
				},
				req: &http.Request{
					Method: http.MethodOptions,
					URL:    mustParseURL("https://sub.domain.com"),
					Header: http.Header{requestHeaderOrigin: {"https://domain.com"}},
				},
			},
			handler: func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
				return response.PlainTextHttpResponseOK("ok"), nil
			},
			wantError: nil,
			wantResponseHeaders: http.Header{
				responseHeaderAccessControlAllowOrigin:  {"https://domain.com"},
				responseHeaderAccessControlAllowMethods: {"options,get"},
				"Content-Type":                          {"text/plain; charset=utf-8"},
				varyHeader:                              {"Origin,Access-Control-Request-Method,Access-Control-Request-Headers"}},
		},
		{
			name: "preFlightCorsRequestType success flow",
			args: args{
				config: CorsConfig{
					AllowedOrigins:     []string{"https://domain.com"},
					AllowedHttpMethods: []string{"options", "get"},
				},
				req: &http.Request{
					Method: http.MethodOptions,
					URL:    mustParseURL("https://sub.domain.com"),
					Header: http.Header{requestHeaderOrigin: {"https://domain.com"}, requestHeaderAccessControlRequestMethod: {"get"}},
				},
			},
			handler: func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
				return response.PlainTextHttpResponseOK("ok"), nil
			},
			wantError: nil,
			wantResponseHeaders: http.Header{
				responseHeaderAccessControlAllowOrigin:  {"https://domain.com"},
				responseHeaderAccessControlAllowMethods: {"options,get"},
				"Content-Type":                          {"text/plain; charset=utf-8"},
				varyHeader:                              {"Origin,Access-Control-Request-Method,Access-Control-Request-Headers"}},
		},
		{
			name: "notCorsRequestType success flow",
			args: args{
				config: CorsConfig{
					AllowedOrigins:     []string{"https://domain.com"},
					AllowedHttpMethods: []string{"options", "get"},
				},
				req: &http.Request{
					Method: http.MethodGet,
					URL:    mustParseURL("https://domain.com"),
					Header: http.Header{requestHeaderOrigin: {"https://domain.com"}},
				},
			},
			handler: func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
				return response.PlainTextHttpResponseOK("ok"), nil
			},
			wantError: nil,
			wantResponseHeaders: http.Header{
				responseHeaderAccessControlAllowOrigin: {"https://domain.com"},
				"Content-Type":                         {"text/plain; charset=utf-8"},
				varyHeader:                             {"Origin"}},
		},
		{
			name: "invalid cors returns error",
			args: args{
				config: CorsConfig{
					AllowedOrigins:     []string{"https://domain.com"},
					AllowedHttpMethods: []string{"options", "post"},
				},
				req: &http.Request{
					Method: http.MethodPost,
					URL:    mustParseURL("https://sub.domain.com"),
					Header: http.Header{requestHeaderOrigin: {"https://domain.com"}, requestHeaderContentType: {""}},
				},
			},
			handler: func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
				return response.PlainTextHttpResponseOK("ok"), nil
			},
			wantError:           errors.ErrDenied,
			wantResponseHeaders: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Cors(tt.args.config)
			resp, err := m(tt.handler)(context.Background(), &request.HttpRequest{R: tt.args.req})
			if tt.wantError != nil {
				if err != tt.wantError {
					t.Errorf("Cors() error = %v, want %v", err, tt.wantError)
				}
			} else if !reflect.DeepEqual(resp.Headers(), tt.wantResponseHeaders) {
				t.Errorf("Cors() response headers = %v, want = %v", resp.Headers(), tt.wantResponseHeaders)
			}
		})
	}
}

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("Failed parsing the url: %s, err:%v", rawURL, err)
	}
	return u
}
