package middleware

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

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
