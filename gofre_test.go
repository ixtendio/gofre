package gofre

import (
	"context"
	"fmt"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/middleware"
	"github.com/ixtendio/gofre/request"
	"github.com/ixtendio/gofre/response"
	"html/template"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"runtime"
	"strings"
	"testing"
	text "text/template"
)

func TestConfig_setDefaults(t *testing.T) {
	tmpl := template.New("")
	errLogFunc := func(err error) {}
	defaultTemplateFunc = func(templatesPathPattern string) (*template.Template, error) { return template.New(""), nil }
	type fields struct {
		ContextPath    string
		TemplateConfig *ResourcesConfig
		ErrLogFunc     func(err error)
	}
	tests := []struct {
		name       string
		fields     fields
		wantErr    bool
		assertFunc func(t *testing.T, c *Config)
	}{
		{
			name:    "test when all values are empty",
			fields:  fields{},
			wantErr: false,
			assertFunc: func(t *testing.T, c *Config) {
				if c.ContextPath != "/" {
					t.Errorf("ContextPath = %v, want /", c.ContextPath)
				}
				if c.ErrLogFunc == nil {
					t.Errorf("ErrLogFunc should not be nil")
				}
				if c.ResourcesConfig != nil {
					t.Errorf("ResourcesConfig should be nil")
				}
			},
		},
		{
			name: "test when ResourcesConfig values are empty",
			fields: fields{
				ContextPath:    "/a_path",
				TemplateConfig: &ResourcesConfig{},
				ErrLogFunc:     errLogFunc,
			},
			wantErr: false,
			assertFunc: func(t *testing.T, c *Config) {
				if c.ContextPath != "/a_path" {
					t.Errorf("ContextPath = \"%s\", want \"\"", c.ContextPath)
				}
				if reflect.ValueOf(c.ErrLogFunc).Pointer() != reflect.ValueOf(errLogFunc).Pointer() {
					t.Errorf("ErrLogFunc should be errLogFunc")
				}
				if c.ResourcesConfig.TemplatesPathPattern != "resources/templates/*.html" {
					t.Errorf("ResourcesConfig.TemplatesPathPattern = %s, want = resources/templates/*.html", c.ResourcesConfig.TemplatesPathPattern)
				}
				if c.ResourcesConfig.AssetsDirPath != "./resources/assets" {
					t.Errorf("ResourcesConfig.AssetsDirPath = %s, want = ./resources/assets", c.ResourcesConfig.AssetsDirPath)
				}
				if c.ResourcesConfig.AssetsMappingPath != "assets" {
					t.Errorf("ResourcesConfig.AssetsMappingPath = %s, want = assets", c.ResourcesConfig.AssetsMappingPath)
				}
				if c.ResourcesConfig.Template == nil {
					t.Errorf("ResourcesConfig.Template is nil, want not nil")
				}
			},
		},
		{
			name: "test when all values are provided",
			fields: fields{
				ContextPath: "/a_path",
				TemplateConfig: &ResourcesConfig{
					TemplatesPathPattern: "pattern",
					AssetsDirPath:        "dir_path",
					AssetsMappingPath:    "assets_path",
					Template:             tmpl,
				},
				ErrLogFunc: errLogFunc,
			},
			wantErr: false,
			assertFunc: func(t *testing.T, c *Config) {
				if c.ContextPath != "/a_path" {
					t.Errorf("ContextPath = \"%s\", want \"\"", c.ContextPath)
				}
				if reflect.ValueOf(c.ErrLogFunc).Pointer() != reflect.ValueOf(errLogFunc).Pointer() {
					t.Errorf("ErrLogFunc should be errLogFunc")
				}
				if c.ResourcesConfig.TemplatesPathPattern != "pattern" {
					t.Errorf("ResourcesConfig.TemplatesPathPattern = %s, want = pattern", c.ResourcesConfig.TemplatesPathPattern)
				}
				if c.ResourcesConfig.AssetsDirPath != "dir_path" {
					t.Errorf("ResourcesConfig.AssetsDirPath = %s, want = dir_path", c.ResourcesConfig.AssetsDirPath)
				}
				if c.ResourcesConfig.AssetsMappingPath != "assets_path" {
					t.Errorf("ResourcesConfig.AssetsMappingPath = %s, want = assets_path", c.ResourcesConfig.AssetsMappingPath)
				}
				if c.ResourcesConfig.Template != tmpl {
					t.Errorf("ResourcesConfig.Template = %v, want = %v", c.ResourcesConfig.Template, tmpl)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				ContextPath:     tt.fields.ContextPath,
				ResourcesConfig: tt.fields.TemplateConfig,
				ErrLogFunc:      tt.fields.ErrLogFunc,
			}
			if err := c.setDefaults(); (err != nil) != tt.wantErr {
				t.Errorf("setDefaults() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.assertFunc(t, c)
		})
	}
}

func TestNewMuxHandlerWithDefaultConfig(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "constructor",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMuxHandlerWithDefaultConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMuxHandlerWithDefaultConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.pathPrefix != "" {
				t.Fatalf("NewMuxHandlerWithDefaultConfig().pathPrefix got: %v, want empty string", got.pathPrefix)
			}
			if got.router == nil {
				t.Fatalf("NewMuxHandlerWithDefaultConfig().router is nil")
			}
			if got.commonMiddlewares != nil {
				t.Fatalf("NewMuxHandlerWithDefaultConfig().commonMiddlewares got: %v, want nil", got.commonMiddlewares)
			}
			if got.webConfig == nil {
				t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig is nil")
			}
			if got.webConfig.ContextPath != "/" {
				t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ContextPath got: %v, want /", got.webConfig.ContextPath)
			}
			if got.webConfig.ErrLogFunc == nil {
				t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ErrLogFunc is nil")
			}
			if got.webConfig.ResourcesConfig != nil {
				t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ResourcesConfig is not nil")
			}
		})
	}
}

func TestNewMuxHandler(t *testing.T) {
	type args struct {
		config *Config
	}
	type want struct {
		pathPrefix               string
		caseInsensitivePathMatch bool
		contextPath              string
		templatesPathPattern     string
		assetsDirPath            string
		assetsMappingPath        string
		htmlTemplate             bool
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "with custom path and case insensitive path match",
			args: args{config: &Config{
				CaseInsensitivePathMatch: true,
				ContextPath:              "/app/",
			}},
			want: want{
				caseInsensitivePathMatch: true,
				contextPath:              "/app/",
			},
			wantErr: false,
		},
		{
			name: "with custom path and case insensitive path match",
			args: args{config: &Config{
				ContextPath: "/",
				ResourcesConfig: &ResourcesConfig{
					TemplatesPathPattern: "TemplatesPathPattern/*html",
					AssetsDirPath:        "AssetsDirPath/",
					AssetsMappingPath:    "AssetsMappingPath",
					Template:             &text.Template{},
				},
			}},
			want: want{
				caseInsensitivePathMatch: false,
				contextPath:              "/",
				templatesPathPattern:     "TemplatesPathPattern/*html",
				assetsDirPath:            "AssetsDirPath/",
				assetsMappingPath:        "AssetsMappingPath",
				htmlTemplate:             false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMuxHandler(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMuxHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.pathPrefix != tt.want.pathPrefix {
				t.Fatalf("NewMuxHandlerWithDefaultConfig().pathPrefix got: %v, want: %v", got.pathPrefix, tt.want.pathPrefix)
			}
			if got.router == nil {
				t.Fatalf("NewMuxHandlerWithDefaultConfig().router is nil")
			}
			if got.commonMiddlewares != nil {
				t.Fatalf("NewMuxHandlerWithDefaultConfig().commonMiddlewares got: %v, want nil", got.commonMiddlewares)
			}
			if got.webConfig == nil {
				t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig is nil")
			}
			if got.webConfig.ContextPath != tt.want.contextPath {
				t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ContextPath got: %v, want: %v", got.webConfig.ContextPath, tt.want.contextPath)
			}
			if got.webConfig.ErrLogFunc == nil {
				t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ErrLogFunc is nil")
			}
			if got.webConfig.ResourcesConfig != nil {
				if got.webConfig.ResourcesConfig.TemplatesPathPattern != tt.want.templatesPathPattern {
					t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ResourcesConfig.TemplatesPathPattern got: %v, want: %v", got.webConfig.ResourcesConfig.TemplatesPathPattern, tt.want.templatesPathPattern)
				}
				if got.webConfig.ResourcesConfig.AssetsDirPath != tt.want.assetsDirPath {
					t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ResourcesConfig.AssetsDirPath got: %v, want: %v", got.webConfig.ResourcesConfig.AssetsDirPath, tt.want.assetsDirPath)
				}
				if got.webConfig.ResourcesConfig.AssetsMappingPath != tt.want.assetsMappingPath {
					t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ResourcesConfig.AssetsMappingPath got: %v, want: %v", got.webConfig.ResourcesConfig.AssetsMappingPath, tt.want.assetsMappingPath)
				}
				if _, ok := got.webConfig.ResourcesConfig.Template.(*template.Template); (tt.want.htmlTemplate && !ok) || (!tt.want.htmlTemplate && ok) {
					t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ResourcesConfig.Template got: %T, want html temlate: %v", got.webConfig.ResourcesConfig.Template, tt.want.htmlTemplate)
				}
			}
		})
	}
}

func TestMuxHandler_RouteWithPathPrefix(t *testing.T) {
	m1 := middleware.PanicRecover()
	m2 := middleware.ErrJsonResponse()
	m3 := middleware.CompressResponse(0)

	type args struct {
		subRouterPath     string
		commonMiddlewares []middleware.Middleware
	}
	type want struct {
		pathPrefix              string
		parentCommonMiddlewares []middleware.Middleware
		childCommonMiddlewares  []middleware.Middleware
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "empty path",
			args: args{
				commonMiddlewares: []middleware.Middleware{m2},
			},
			want: want{
				pathPrefix:              "/",
				parentCommonMiddlewares: []middleware.Middleware{m1, m2},
				childCommonMiddlewares:  []middleware.Middleware{m1, m2},
			},
		},
		{
			name: "new path",
			args: args{
				subRouterPath:     "/users",
				commonMiddlewares: []middleware.Middleware{m2, m3},
			},
			want: want{
				pathPrefix:              "/users",
				parentCommonMiddlewares: []middleware.Middleware{m1},
				childCommonMiddlewares:  []middleware.Middleware{m1, m2, m3},
			},
		},
	}
	for _, tt := range tests {
		parent, _ := NewMuxHandlerWithDefaultConfig()
		parent.CommonMiddlewares(m1)
		t.Run(tt.name, func(t *testing.T) {
			child := parent.RouteWithPathPrefix(tt.args.subRouterPath)
			child.CommonMiddlewares(tt.args.commonMiddlewares...)

			if err := compareMiddlewares(parent.commonMiddlewares, tt.want.parentCommonMiddlewares); err != nil {
				t.Fatalf("RouteWithPathPrefix() parent.commonMiddlewares: %v", err)
			}
			if err := compareMiddlewares(child.commonMiddlewares, tt.want.childCommonMiddlewares); err != nil {
				t.Fatalf("RouteWithPathPrefix() child.commonMiddlewares: %v", err)
			}
		})
	}
}

func compareMiddlewares(m1 []middleware.Middleware, m2 []middleware.Middleware) error {
	if len(m1) != len(m2) {
		return fmt.Errorf("length => got: %v, want: %v", len(m1), len(m2))
	}
	for i := 0; i < len(m1); i++ {
		m1FuncName := runtime.FuncForPC(reflect.ValueOf(m1[i]).Pointer()).Name()
		m2FuncName := runtime.FuncForPC(reflect.ValueOf(m2[i]).Pointer()).Name()
		if m1FuncName != m2FuncName {
			return fmt.Errorf("istance => index: %v, got: %v, want: %v", i, m1FuncName, m2FuncName)
		}
	}
	return nil
}

func TestMuxHandler_resolvePath(t *testing.T) {
	type args struct {
		currentPath string
		newPath     string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty newPath",
			args: args{
				currentPath: "/",
				newPath:     "",
			},
			want: "/",
		},
		{
			name: "empty currentPath",
			args: args{
				currentPath: "",
				newPath:     "/",
			},
			want: "/",
		},
		{
			name: "currentPath and newPath are /",
			args: args{
				currentPath: "/",
				newPath:     "/",
			},
			want: "/",
		},
		{
			name: "currentPath ends with /",
			args: args{
				currentPath: "/a/",
				newPath:     "b",
			},
			want: "/a/b",
		},
		{
			name: "currentPath ends with /, newPath starts with /",
			args: args{
				currentPath: "/a/",
				newPath:     "/b",
			},
			want: "/a/b",
		},
		{
			name: "newPath starts with /",
			args: args{
				currentPath: "/a",
				newPath:     "/b",
			},
			want: "/a/b",
		},
		{
			name: "separator should be added",
			args: args{
				currentPath: "/a",
				newPath:     "b",
			},
			want: "/a/b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := NewMuxHandlerWithDefaultConfig()
			m.pathPrefix = tt.args.currentPath
			if got := m.resolvePath(tt.args.newPath); got != tt.want {
				t.Errorf("resolvePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMuxHandler_HandleRequest(t *testing.T) {
	var sb strings.Builder
	type args struct {
		httpMethod          string
		path                string
		h                   handler.Handler
		commonMiddlewares   []middleware.Middleware
		endpointMiddlewares []middleware.Middleware
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "register commonMiddlewares and endpoint middlewares",
			args: args{
				httpMethod: "GET",
				path:       "/test",
				h: func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
					sb.WriteString("handler")
					return response.PlainTextHttpResponseOK(""), nil
				},
				commonMiddlewares: []middleware.Middleware{
					func(handler handler.Handler) handler.Handler {
						return func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
							sb.WriteString("com1:")
							return handler(ctx, r)
						}
					}, func(handler handler.Handler) handler.Handler {
						return func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
							sb.WriteString("com2:")
							return handler(ctx, r)
						}
					},
				},
				endpointMiddlewares: []middleware.Middleware{
					func(handler handler.Handler) handler.Handler {
						return func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
							sb.WriteString("cust1:")
							return handler(ctx, r)
						}
					}, func(handler handler.Handler) handler.Handler {
						return func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
							sb.WriteString("cust2:")
							return handler(ctx, r)
						}
					},
				},
			},
			want: "com1:com2:cust1:cust2:handler",
		},
	}
	for _, tt := range tests {
		sb.Reset()
		responseRecorder := httptest.NewRecorder()
		t.Run(tt.name, func(t *testing.T) {
			m, _ := NewMuxHandlerWithDefaultConfig()
			m.CommonMiddlewares(tt.args.commonMiddlewares...)
			m.HandleRequest(tt.args.httpMethod, tt.args.path, tt.args.h, tt.args.endpointMiddlewares...)

			m.ServeHTTP(responseRecorder, &http.Request{Method: tt.args.httpMethod, URL: mustParseURL("https://domain.com" + tt.args.path)})
			if responseRecorder.Code != 200 {
				t.Errorf("ServeHTTP() got responseCode: %v, want: 200", responseRecorder.Code)
			}
			if sb.String() != tt.want {
				t.Errorf("ServeHTTP() got: %v, want: %v", sb.String(), tt.want)
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
