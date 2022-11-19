package gofre

//
//import (
//	"context"
//	"fmt"
//	"github.com/ixtendio/gofre/auth"
//	"github.com/ixtendio/gofre/auth/oauth"
//	"github.com/ixtendio/gofre/cache"
//	"github.com/ixtendio/gofre/handler"
//	"github.com/ixtendio/gofre/middleware"
//
//	"github.com/ixtendio/gofre/response"
//	"html/template"
//	"log"
//	"net/http"
//	"net/http/httptest"
//	"net/url"
//	"reflect"
//	"runtime"
//	"strconv"
//	"strings"
//	"testing"
//	text "text/template"
//	"time"
//)
//
//type captureOAUTHProvider struct {
//	name string
//}
//
//func (p captureOAUTHProvider) Name() string {
//	return p.name
//}
//
//func (p captureOAUTHProvider) InitiateUrl(redirectUri string, state string, includeUserInfoProfileScope bool) string {
//	return redirectUri + "?" + "state=" + state + "&profile=" + strconv.FormatBool(includeUserInfoProfileScope)
//}
//
//func (p captureOAUTHProvider) FetchAccessToken(ctx context.Context, redirectUri string, authCode string) (oauth.AccessToken, error) {
//	return oauth.AccessToken{
//		AccessToken: redirectUri + ":" + authCode,
//	}, nil
//}
//
//func (p captureOAUTHProvider) FetchAuthenticatedUser(ctx context.Context, accessToken oauth.AccessToken) (auth.User, error) {
//	return auth.User{Id: "user123"}, nil
//}
//
//func TestConfig_setDefaults(t *testing.T) {
//	tmpl := template.New("")
//	errLogFunc := func(err error) {}
//	defaultTemplateFunc = func(templatesPathPattern string) (*template.Template, error) { return template.New(""), nil }
//	type fields struct {
//		ContextPath    string
//		TemplateConfig *ResourcesConfig
//		ErrLogFunc     func(err error)
//	}
//	tests := []struct {
//		name       string
//		fields     fields
//		wantErr    bool
//		assertFunc func(t *testing.T, c *Config)
//	}{
//		{
//			name:    "test when all values are empty",
//			fields:  fields{},
//			wantErr: false,
//			assertFunc: func(t *testing.T, c *Config) {
//				if c.ContextPath != "/" {
//					t.Errorf("ContextPath = %v, want /", c.ContextPath)
//				}
//				if c.ErrLogFunc == nil {
//					t.Errorf("ErrLogFunc should not be nil")
//				}
//				if c.ResourcesConfig != nil {
//					t.Errorf("ResourcesConfig should be nil")
//				}
//			},
//		},
//		{
//			name: "test when ResourcesConfig values are empty",
//			fields: fields{
//				ContextPath:    "/a_path",
//				TemplateConfig: &ResourcesConfig{},
//				ErrLogFunc:     errLogFunc,
//			},
//			wantErr: false,
//			assertFunc: func(t *testing.T, c *Config) {
//				if c.ContextPath != "/a_path" {
//					t.Errorf("ContextPath = \"%s\", want \"\"", c.ContextPath)
//				}
//				if reflect.ValueOf(c.ErrLogFunc).Pointer() != reflect.ValueOf(errLogFunc).Pointer() {
//					t.Errorf("ErrLogFunc should be errLogFunc")
//				}
//				if c.ResourcesConfig.TemplatesPathPattern != "resources/templates/*.html" {
//					t.Errorf("ResourcesConfig.TemplatesPathPattern = %s, want = resources/templates/*.html", c.ResourcesConfig.TemplatesPathPattern)
//				}
//				if c.ResourcesConfig.AssetsDirPath != "./resources/assets" {
//					t.Errorf("ResourcesConfig.AssetsDirPath = %s, want = ./resources/assets", c.ResourcesConfig.AssetsDirPath)
//				}
//				if c.ResourcesConfig.AssetsMappingPath != "assets" {
//					t.Errorf("ResourcesConfig.AssetsMappingPath = %s, want = assets", c.ResourcesConfig.AssetsMappingPath)
//				}
//				if c.ResourcesConfig.Template == nil {
//					t.Errorf("ResourcesConfig.Template is nil, want not nil")
//				}
//			},
//		},
//		{
//			name: "test when all values are provided",
//			fields: fields{
//				ContextPath: "/a_path",
//				TemplateConfig: &ResourcesConfig{
//					TemplatesPathPattern: "pattern",
//					AssetsDirPath:        "dir_path",
//					AssetsMappingPath:    "assets_path",
//					Template:             tmpl,
//				},
//				ErrLogFunc: errLogFunc,
//			},
//			wantErr: false,
//			assertFunc: func(t *testing.T, c *Config) {
//				if c.ContextPath != "/a_path" {
//					t.Errorf("ContextPath = \"%s\", want \"\"", c.ContextPath)
//				}
//				if reflect.ValueOf(c.ErrLogFunc).Pointer() != reflect.ValueOf(errLogFunc).Pointer() {
//					t.Errorf("ErrLogFunc should be errLogFunc")
//				}
//				if c.ResourcesConfig.TemplatesPathPattern != "pattern" {
//					t.Errorf("ResourcesConfig.TemplatesPathPattern = %s, want = pattern", c.ResourcesConfig.TemplatesPathPattern)
//				}
//				if c.ResourcesConfig.AssetsDirPath != "dir_path" {
//					t.Errorf("ResourcesConfig.AssetsDirPath = %s, want = dir_path", c.ResourcesConfig.AssetsDirPath)
//				}
//				if c.ResourcesConfig.AssetsMappingPath != "assets_path" {
//					t.Errorf("ResourcesConfig.AssetsMappingPath = %s, want = assets_path", c.ResourcesConfig.AssetsMappingPath)
//				}
//				if c.ResourcesConfig.Template != tmpl {
//					t.Errorf("ResourcesConfig.Template = %v, want = %v", c.ResourcesConfig.Template, tmpl)
//				}
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c := &Config{
//				ContextPath:     tt.fields.ContextPath,
//				ResourcesConfig: tt.fields.TemplateConfig,
//				ErrLogFunc:      tt.fields.ErrLogFunc,
//			}
//			if err := c.setDefaults(); (err != nil) != tt.wantErr {
//				t.Errorf("setDefaults() error = %v, wantErr %v", err, tt.wantErr)
//			}
//			tt.assertFunc(t, c)
//		})
//	}
//}
//
//func TestNewMuxHandlerWithDefaultConfig(t *testing.T) {
//	tests := []struct {
//		name    string
//		wantErr bool
//	}{
//		{
//			name:    "constructor",
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := NewMuxHandlerWithDefaultConfig()
//			if (err != nil) != tt.wantErr {
//				t.Errorf("NewMuxHandlerWithDefaultConfig() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if got.pathPrefix != "" {
//				t.Fatalf("NewMuxHandlerWithDefaultConfig().pathPrefix got: %v, want empty string", got.pathPrefix)
//			}
//			if got.router == nil {
//				t.Fatalf("NewMuxHandlerWithDefaultConfig().router is nil")
//			}
//			if got.commonMiddlewares != nil {
//				t.Fatalf("NewMuxHandlerWithDefaultConfig().childCommonMiddlewares got: %v, want nil", got.commonMiddlewares)
//			}
//			if got.webConfig == nil {
//				t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig is nil")
//			}
//			if got.webConfig.ContextPath != "/" {
//				t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ContextPath got: %v, want /", got.webConfig.ContextPath)
//			}
//			if got.webConfig.ErrLogFunc == nil {
//				t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ErrLogFunc is nil")
//			}
//			if got.webConfig.ResourcesConfig != nil {
//				t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ResourcesConfig is not nil")
//			}
//		})
//	}
//}
//
//func TestNewMuxHandlerWithDefaultConfigAndTemplateSupport(t *testing.T) {
//	defaultTemplateFunc = func(templatesPathPattern string) (*template.Template, error) {
//		return &template.Template{}, nil
//	}
//	tests := []struct {
//		name    string
//		want    *MuxHandler
//		wantErr bool
//	}{
//		{
//			name:    "constructor",
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := NewMuxHandlerWithDefaultConfigAndTemplateSupport()
//			if (err != nil) != tt.wantErr {
//				t.Errorf("NewMuxHandlerWithDefaultConfigAndTemplateSupport() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if got.pathPrefix != "" {
//				t.Fatalf("NewMuxHandlerWithDefaultConfigAndTemplateSupport().pathPrefix got: %v, want empty string", got.pathPrefix)
//			}
//			if got.router == nil {
//				t.Fatalf("NewMuxHandlerWithDefaultConfigAndTemplateSupport().router is nil")
//			}
//			if got.commonMiddlewares != nil {
//				t.Fatalf("NewMuxHandlerWithDefaultConfigAndTemplateSupport().childCommonMiddlewares got: %v, want nil", got.commonMiddlewares)
//			}
//			if got.webConfig == nil {
//				t.Fatalf("NewMuxHandlerWithDefaultConfigAndTemplateSupport().webConfig is nil")
//			}
//			if got.webConfig.ContextPath != "/" {
//				t.Fatalf("NewMuxHandlerWithDefaultConfigAndTemplateSupport().webConfig.ContextPath got: %v, want /", got.webConfig.ContextPath)
//			}
//			if got.webConfig.ErrLogFunc == nil {
//				t.Fatalf("NewMuxHandlerWithDefaultConfigAndTemplateSupport().webConfig.ErrLogFunc is nil")
//			}
//			if got.webConfig.ResourcesConfig == nil {
//				t.Fatalf("NewMuxHandlerWithDefaultConfigAndTemplateSupport().webConfig.ResourcesConfig is nil")
//			}
//			if got.webConfig.ResourcesConfig.TemplatesPathPattern != "resources/templates/*.html" {
//				t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ResourcesConfig.TemplatesPathPattern got: %v, want: %v", got.webConfig.ResourcesConfig.TemplatesPathPattern, "resources/templates/*.html")
//			}
//			if got.webConfig.ResourcesConfig.AssetsDirPath != "./resources/assets" {
//				t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ResourcesConfig.AssetsDirPath got: %v, want: %v", got.webConfig.ResourcesConfig.AssetsDirPath, "./resources/assets")
//			}
//			if got.webConfig.ResourcesConfig.AssetsMappingPath != "assets" {
//				t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ResourcesConfig.AssetsMappingPath got: %v, want: %v", got.webConfig.ResourcesConfig.AssetsMappingPath, "assets")
//			}
//			if got.webConfig.ResourcesConfig.Template == nil {
//				t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ResourcesConfig.Template got: nil, want: not nil")
//			}
//		})
//	}
//}
//
//func TestNewDefaultResourcesConfig(t *testing.T) {
//	defaultTemplateFunc = func(templatesPathPattern string) (*template.Template, error) {
//		return &template.Template{}, nil
//	}
//	tests := []struct {
//		name string
//	}{
//		{name: "constructor"},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got := NewDefaultResourcesConfig()
//			if got.TemplatesPathPattern != "resources/templates/*.html" {
//				t.Fatalf("NewDefaultResourcesConfig().TemplatesPathPattern got: %v, want: %v", got.TemplatesPathPattern, "resources/templates/*.html")
//			}
//			if got.AssetsDirPath != "./resources/assets" {
//				t.Fatalf("NewDefaultResourcesConfig().AssetsDirPath got: %v, want: %v", got.AssetsDirPath, "./resources/assets")
//			}
//			if got.AssetsMappingPath != "assets" {
//				t.Fatalf("NewDefaultResourcesConfig().AssetsMappingPath got: %v, want: %v", got.AssetsMappingPath, "assets")
//			}
//			if got.Template == nil {
//				t.Fatalf("NewDefaultResourcesConfig().Template got: nil, want: not nil")
//			}
//		})
//	}
//}
//
//func TestNewMuxHandler(t *testing.T) {
//	type args struct {
//		config *Config
//	}
//	type want struct {
//		pathPrefix               string
//		caseInsensitivePathMatch bool
//		contextPath              string
//		templatesPathPattern     string
//		assetsDirPath            string
//		assetsMappingPath        string
//		htmlTemplate             bool
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    want
//		wantErr bool
//	}{
//		{
//			name: "with custom path and case insensitive path match",
//			args: args{config: &Config{
//				CaseInsensitivePathMatch: true,
//				ContextPath:              "/app/",
//			}},
//			want: want{
//				caseInsensitivePathMatch: true,
//				contextPath:              "/app/",
//			},
//			wantErr: false,
//		},
//		{
//			name: "with custom path and case insensitive path match",
//			args: args{config: &Config{
//				ContextPath: "/",
//				ResourcesConfig: &ResourcesConfig{
//					TemplatesPathPattern: "TemplatesPathPattern/*html",
//					AssetsDirPath:        "AssetsDirPath/",
//					AssetsMappingPath:    "AssetsMappingPath",
//					Template:             &text.Template{},
//				},
//			}},
//			want: want{
//				caseInsensitivePathMatch: false,
//				contextPath:              "/",
//				templatesPathPattern:     "TemplatesPathPattern/*html",
//				assetsDirPath:            "AssetsDirPath/",
//				assetsMappingPath:        "AssetsMappingPath",
//				htmlTemplate:             false,
//			},
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := NewMuxHandler(tt.args.config)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("NewMuxHandler() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if got.pathPrefix != tt.want.pathPrefix {
//				t.Fatalf("NewMuxHandlerWithDefaultConfig().pathPrefix got: %v, want: %v", got.pathPrefix, tt.want.pathPrefix)
//			}
//			if got.router == nil {
//				t.Fatalf("NewMuxHandlerWithDefaultConfig().router is nil")
//			}
//			if got.commonMiddlewares != nil {
//				t.Fatalf("NewMuxHandlerWithDefaultConfig().childCommonMiddlewares got: %v, want nil", got.commonMiddlewares)
//			}
//			if got.webConfig == nil {
//				t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig is nil")
//			}
//			if got.webConfig.ContextPath != tt.want.contextPath {
//				t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ContextPath got: %v, want: %v", got.webConfig.ContextPath, tt.want.contextPath)
//			}
//			if got.webConfig.ErrLogFunc == nil {
//				t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ErrLogFunc is nil")
//			}
//			if got.webConfig.ResourcesConfig != nil {
//				if got.webConfig.ResourcesConfig.TemplatesPathPattern != tt.want.templatesPathPattern {
//					t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ResourcesConfig.TemplatesPathPattern got: %v, want: %v", got.webConfig.ResourcesConfig.TemplatesPathPattern, tt.want.templatesPathPattern)
//				}
//				if got.webConfig.ResourcesConfig.AssetsDirPath != tt.want.assetsDirPath {
//					t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ResourcesConfig.AssetsDirPath got: %v, want: %v", got.webConfig.ResourcesConfig.AssetsDirPath, tt.want.assetsDirPath)
//				}
//				if got.webConfig.ResourcesConfig.AssetsMappingPath != tt.want.assetsMappingPath {
//					t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ResourcesConfig.AssetsMappingPath got: %v, want: %v", got.webConfig.ResourcesConfig.AssetsMappingPath, tt.want.assetsMappingPath)
//				}
//				if _, ok := got.webConfig.ResourcesConfig.Template.(*template.Template); (tt.want.htmlTemplate && !ok) || (!tt.want.htmlTemplate && ok) {
//					t.Fatalf("NewMuxHandlerWithDefaultConfig().webConfig.ResourcesConfig.Template got: %T, want html temlate: %v", got.webConfig.ResourcesConfig.Template, tt.want.htmlTemplate)
//				}
//			}
//		})
//	}
//}
//
//func TestMuxHandler_Config(t *testing.T) {
//	tests := []struct {
//		name string
//	}{
//		{name: "function call"},
//	}
//
//	checkConfig := func(t *testing.T, cfg Config) {
//		if cfg.ContextPath != "/" {
//			t.Fatalf("Config().ContextPath got: %v, want /", cfg.ContextPath)
//		}
//		if cfg.ErrLogFunc == nil {
//			t.Fatalf("Config().ErrLogFunc is nil")
//		}
//		if cfg.ResourcesConfig != nil {
//			t.Fatalf("Config().ResourcesConfig is not nil")
//		}
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			m, _ := NewMuxHandlerWithDefaultConfig()
//			got := m.Config()
//			checkConfig(t, got)
//
//			got.ContextPath = "/test"
//			got.ResourcesConfig = &ResourcesConfig{}
//
//			// check for mutations
//			checkConfig(t, m.Config())
//		})
//	}
//}
//
//func TestMuxHandler_ExecutableTemplate(t *testing.T) {
//	tmpl := &template.Template{}
//	defaultTemplateFunc = func(templatesPathPattern string) (*template.Template, error) {
//		return tmpl, nil
//	}
//	type args struct {
//		templateSupport bool
//	}
//	tests := []struct {
//		name string
//		args args
//	}{
//		{
//			name: "with template support",
//			args: args{templateSupport: true},
//		},
//		{
//			name: "without template support",
//			args: args{templateSupport: false},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if tt.args.templateSupport {
//				m, _ := NewMuxHandlerWithDefaultConfigAndTemplateSupport()
//				et := m.ExecutableTemplate()
//				if et != tmpl {
//					t.Fatalf("ExecutableTemplate() with template support got: %v, want: %v", et, tmpl)
//				}
//			} else {
//				m, _ := NewMuxHandlerWithDefaultConfig()
//				et := m.ExecutableTemplate()
//				if et != nil {
//					t.Fatalf("ExecutableTemplate() without template support got: %v, want nil", et)
//				}
//			}
//		})
//	}
//}
//
//func TestMuxHandler_Clone(t *testing.T) {
//	m1 := middleware.PanicRecover()
//	m2 := middleware.ErrJsonResponse()
//	m3 := middleware.CompressResponse(0)
//
//	type args struct {
//		parentCommonMiddlewares []middleware.Middleware
//		childCommonMiddlewares  []middleware.Middleware
//	}
//	type want struct {
//		pathPrefix              string
//		parentCommonMiddlewares []middleware.Middleware
//		childCommonMiddlewares  []middleware.Middleware
//	}
//	tests := []struct {
//		name string
//		args args
//		want want
//	}{
//		{
//			name: "clone",
//			args: args{
//				parentCommonMiddlewares: []middleware.Middleware{m1},
//				childCommonMiddlewares:  []middleware.Middleware{m2, m3},
//			},
//			want: want{
//				parentCommonMiddlewares: []middleware.Middleware{m1},
//				childCommonMiddlewares:  []middleware.Middleware{m1, m2, m3},
//			},
//		},
//		{
//			name: "clone using nil common middlewares",
//			args: args{},
//			want: want{},
//		},
//	}
//	for _, tt := range tests {
//		parent, _ := NewMuxHandlerWithDefaultConfig()
//		parent.CommonMiddlewares(tt.args.parentCommonMiddlewares...)
//		t.Run(tt.name, func(t *testing.T) {
//			child := parent.Clone()
//			child.CommonMiddlewares(tt.args.childCommonMiddlewares...)
//
//			if err := compareMiddlewares(parent.commonMiddlewares, tt.want.parentCommonMiddlewares); err != nil {
//				t.Fatalf("Clone() parent.childCommonMiddlewares: %v", err)
//			}
//			if err := compareMiddlewares(child.commonMiddlewares, tt.want.childCommonMiddlewares); err != nil {
//				t.Fatalf("Clone() child.childCommonMiddlewares: %v", err)
//			}
//			if !reflect.DeepEqual(parent.pathPrefix, child.pathPrefix) {
//				t.Fatalf("Clone() parent.pathPrefix: %v child.pathPrefix: %v", parent.pathPrefix, child.pathPrefix)
//			}
//			if !reflect.DeepEqual(parent.router, child.router) {
//				t.Fatalf("Clone() parent.router: %v child.router: %v", parent.router, child.router)
//			}
//			if !reflect.DeepEqual(parent.webConfig, child.webConfig) {
//				t.Fatalf("Clone() parent.webConfig: %v child.webConfig: %v", parent.webConfig, child.webConfig)
//			}
//		})
//	}
//}
//
//func TestMuxHandler_RouteWithPathPrefix(t *testing.T) {
//	m1 := middleware.PanicRecover()
//	m2 := middleware.ErrJsonResponse()
//	m3 := middleware.CompressResponse(0)
//
//	type args struct {
//		subRouterPath     string
//		commonMiddlewares []middleware.Middleware
//	}
//	type want struct {
//		pathPrefix              string
//		parentCommonMiddlewares []middleware.Middleware
//		childCommonMiddlewares  []middleware.Middleware
//	}
//	tests := []struct {
//		name string
//		args args
//		want want
//	}{
//		{
//			name: "empty path",
//			args: args{
//				commonMiddlewares: []middleware.Middleware{m2},
//			},
//			want: want{
//				pathPrefix:              "",
//				parentCommonMiddlewares: []middleware.Middleware{m1, m2},
//				childCommonMiddlewares:  []middleware.Middleware{m1, m2},
//			},
//		},
//		{
//			name: "new path",
//			args: args{
//				subRouterPath:     "/users",
//				commonMiddlewares: []middleware.Middleware{m2, m3},
//			},
//			want: want{
//				pathPrefix:              "/users",
//				parentCommonMiddlewares: []middleware.Middleware{m1},
//				childCommonMiddlewares:  []middleware.Middleware{m1, m2, m3},
//			},
//		},
//	}
//	for _, tt := range tests {
//		parent, _ := NewMuxHandlerWithDefaultConfig()
//		parent.CommonMiddlewares(m1)
//		t.Run(tt.name, func(t *testing.T) {
//			child := parent.RouteUsingPathPrefix(tt.args.subRouterPath)
//			child.CommonMiddlewares(tt.args.commonMiddlewares...)
//
//			if err := compareMiddlewares(parent.commonMiddlewares, tt.want.parentCommonMiddlewares); err != nil {
//				t.Fatalf("RouteUsingPathPrefix() parent.childCommonMiddlewares: %v", err)
//			}
//			if err := compareMiddlewares(child.commonMiddlewares, tt.want.childCommonMiddlewares); err != nil {
//				t.Fatalf("RouteUsingPathPrefix() child.childCommonMiddlewares: %v", err)
//			}
//			if !reflect.DeepEqual(tt.want.pathPrefix, child.pathPrefix) {
//				t.Fatalf("RouteUsingPathPrefix() want.pathPrefix: %v child.pathPrefix: %v", tt.want.pathPrefix, child.pathPrefix)
//			}
//			if !reflect.DeepEqual(parent.router, child.router) {
//				t.Fatalf("RouteUsingPathPrefix() parent.router: %v child.router: %v", parent.router, child.router)
//			}
//			if !reflect.DeepEqual(parent.webConfig, child.webConfig) {
//				t.Fatalf("RouteUsingPathPrefix() parent.webConfig: %v child.webConfig: %v", parent.webConfig, child.webConfig)
//			}
//		})
//	}
//}
//
//func TestMuxHandler_resolvePath(t *testing.T) {
//	type args struct {
//		currentPath string
//		newPath     string
//	}
//	tests := []struct {
//		name string
//		args args
//		want string
//	}{
//		{
//			name: "empty newPath",
//			args: args{
//				currentPath: "/",
//				newPath:     "",
//			},
//			want: "/",
//		},
//		{
//			name: "empty currentPath",
//			args: args{
//				currentPath: "",
//				newPath:     "/",
//			},
//			want: "/",
//		},
//		{
//			name: "currentPath and newPath are /",
//			args: args{
//				currentPath: "/",
//				newPath:     "/",
//			},
//			want: "/",
//		},
//		{
//			name: "currentPath ends with /",
//			args: args{
//				currentPath: "/a/",
//				newPath:     "b",
//			},
//			want: "/a/b",
//		},
//		{
//			name: "currentPath ends with /, newPath starts with /",
//			args: args{
//				currentPath: "/a/",
//				newPath:     "/b",
//			},
//			want: "/a/b",
//		},
//		{
//			name: "newPath starts with /",
//			args: args{
//				currentPath: "/a",
//				newPath:     "/b",
//			},
//			want: "/a/b",
//		},
//		{
//			name: "separator should be added",
//			args: args{
//				currentPath: "/a",
//				newPath:     "b",
//			},
//			want: "/a/b",
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			m, _ := NewMuxHandlerWithDefaultConfig()
//			m.pathPrefix = tt.args.currentPath
//			if got := m.resolvePath(tt.args.newPath); got != tt.want {
//				t.Errorf("resolvePath() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestMuxHandler_HandleRequest(t *testing.T) {
//	var sb strings.Builder
//	type args struct {
//		httpMethod          string
//		path                string
//		h                   handler.Handler
//		commonMiddlewares   []middleware.Middleware
//		endpointMiddlewares []middleware.Middleware
//	}
//	tests := []struct {
//		name string
//		args args
//		want string
//	}{
//		{
//			name: "register childCommonMiddlewares and endpoint middlewares",
//			args: args{
//				httpMethod: "GET",
//				path:       "/test",
//				h: func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
//					sb.WriteString("handler")
//					return response.PlainTextHttpResponseOK(""), nil
//				},
//				commonMiddlewares: []middleware.Middleware{
//					func(handler handler.Handler) handler.Handler {
//						return func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
//							sb.WriteString("com1:")
//							return handler(ctx, r)
//						}
//					}, func(handler handler.Handler) handler.Handler {
//						return func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
//							sb.WriteString("com2:")
//							return handler(ctx, r)
//						}
//					},
//				},
//				endpointMiddlewares: []middleware.Middleware{
//					func(handler handler.Handler) handler.Handler {
//						return func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
//							sb.WriteString("cust1:")
//							return handler(ctx, r)
//						}
//					}, func(handler handler.Handler) handler.Handler {
//						return func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
//							sb.WriteString("cust2:")
//							return handler(ctx, r)
//						}
//					},
//				},
//			},
//			want: "com1:com2:cust1:cust2:handler",
//		},
//	}
//	for _, tt := range tests {
//		sb.Reset()
//		responseRecorder := httptest.NewRecorder()
//		t.Run(tt.name, func(t *testing.T) {
//			m, _ := NewMuxHandlerWithDefaultConfig()
//			m.CommonMiddlewares(tt.args.commonMiddlewares...)
//			m.HandleRequest(tt.args.httpMethod, tt.args.path, tt.args.h, tt.args.endpointMiddlewares...)
//
//			m.ServeHTTP(responseRecorder, &http.Request{Method: tt.args.httpMethod, URL: mustParseURL("https://domain.com" + tt.args.path)})
//			if responseRecorder.Code != 200 {
//				t.Errorf("ServeHTTP() got responseCode: %v, want: 200", responseRecorder.Code)
//			}
//			if sb.String() != tt.want {
//				t.Errorf("ServeHTTP() got: %v, want: %v", sb.String(), tt.want)
//			}
//		})
//	}
//}
//
//func TestMuxHandler_HandleXXX(t *testing.T) {
//	var sb strings.Builder
//	type args struct {
//		httpMethod string
//		path       string
//	}
//	tests := []struct {
//		name             string
//		argsSupplierFunc func(m *MuxHandler) args
//		want             string
//	}{
//		{
//			name: "GET",
//			argsSupplierFunc: func(m *MuxHandler) args {
//				m.HandleGet("/get", func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
//					sb.WriteString("get_handler")
//					return response.PlainTextHttpResponseOK(""), nil
//				}, func(handler handler.Handler) handler.Handler {
//					return func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
//						sb.WriteString("get_middleware:")
//						return handler(ctx, r)
//					}
//				})
//				return args{
//					httpMethod: "GET",
//					path:       "/get",
//				}
//			},
//			want: "get_middleware:get_handler",
//		},
//		{
//			name: "POST",
//			argsSupplierFunc: func(m *MuxHandler) args {
//				m.HandlePost("/post", func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
//					sb.WriteString("post_handler")
//					return response.PlainTextHttpResponseOK(""), nil
//				}, func(handler handler.Handler) handler.Handler {
//					return func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
//						sb.WriteString("post_middleware:")
//						return handler(ctx, r)
//					}
//				})
//				return args{
//					httpMethod: "POST",
//					path:       "/post",
//				}
//			},
//			want: "post_middleware:post_handler",
//		},
//		{
//			name: "PUT",
//			argsSupplierFunc: func(m *MuxHandler) args {
//				m.HandlePut("/put", func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
//					sb.WriteString("put_handler")
//					return response.PlainTextHttpResponseOK(""), nil
//				}, func(handler handler.Handler) handler.Handler {
//					return func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
//						sb.WriteString("put_middleware:")
//						return handler(ctx, r)
//					}
//				})
//				return args{
//					httpMethod: "PUT",
//					path:       "/put",
//				}
//			},
//			want: "put_middleware:put_handler",
//		}, {
//			name: "PATCH",
//			argsSupplierFunc: func(m *MuxHandler) args {
//				m.HandlePatch("/patch", func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
//					sb.WriteString("patch_handler")
//					return response.PlainTextHttpResponseOK(""), nil
//				}, func(handler handler.Handler) handler.Handler {
//					return func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
//						sb.WriteString("patch_middleware:")
//						return handler(ctx, r)
//					}
//				})
//				return args{
//					httpMethod: "PATCH",
//					path:       "/patch",
//				}
//			},
//			want: "patch_middleware:patch_handler",
//		}, {
//			name: "DELETE",
//			argsSupplierFunc: func(m *MuxHandler) args {
//				m.HandleDelete("/delete", func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
//					sb.WriteString("delete_handler")
//					return response.PlainTextHttpResponseOK(""), nil
//				}, func(handler handler.Handler) handler.Handler {
//					return func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
//						sb.WriteString("delete_middleware:")
//						return handler(ctx, r)
//					}
//				})
//				return args{
//					httpMethod: "DELETE",
//					path:       "/delete",
//				}
//			},
//			want: "delete_middleware:delete_handler",
//		},
//	}
//	for _, tt := range tests {
//		sb.Reset()
//		responseRecorder := httptest.NewRecorder()
//		t.Run(tt.name, func(t *testing.T) {
//			m, _ := NewMuxHandlerWithDefaultConfig()
//			args := tt.argsSupplierFunc(m)
//
//			m.ServeHTTP(responseRecorder, &http.Request{Method: args.httpMethod, URL: mustParseURL("https://domain.com" + args.path)})
//			if responseRecorder.Code != 200 {
//				t.Errorf("HandleXXX() got responseCode: %v, want: 200", responseRecorder.Code)
//			}
//			if sb.String() != tt.want {
//				t.Errorf("HandleXXX() got: %v, want: %v", sb.String(), tt.want)
//			}
//		})
//	}
//}
//
//func TestMuxHandler_HandleOAUTH2(t *testing.T) {
//	var sb strings.Builder
//	uniqueIdFunc = func(length int) string {
//		return "123456"
//	}
//	type args struct {
//		providers        []oauth.Provider
//		websiteUrl       string
//		initiateURL      string
//		authorizeURL     string
//		fetchUserDetails bool
//	}
//	type want struct {
//		initiateResponse    string
//		authorizeResponse   string
//		initiateStatusCode  int
//		authorizeStatusCode int
//	}
//	tests := []struct {
//		name string
//		args args
//		want want
//	}{
//		{
//			name: "using a singleRoute provider and fetching user details",
//			args: args{
//				providers:        []oauth.Provider{captureOAUTHProvider{name: "ixtendio"}},
//				websiteUrl:       "https://www.domain.com",
//				initiateURL:      "https://www.domain.com/oauth/initiate",
//				authorizeURL:     "https://www.domain.com/oauth/authorize/ixtendio?state=123456&code=code123",
//				fetchUserDetails: true,
//			},
//			want: want{
//				initiateResponse:    `<a href="https://www.domain.com/oauth/authorize/ixtendio?state=123456&amp;profile=true">Found</a>`,
//				authorizeResponse:   "initiateOAUTHMiddleware|authorizeOAUTHMiddleware|at=https://www.domain.com/oauth/authorize/ixtendio:code123;sp=user123",
//				initiateStatusCode:  http.StatusFound,
//				authorizeStatusCode: 200,
//			},
//		},
//		{
//			name: "using 2 providers provider",
//			args: args{
//				providers:        []oauth.Provider{captureOAUTHProvider{name: "ixtendio"}, captureOAUTHProvider{name: "ixtendio1"}},
//				websiteUrl:       "https://www.domain.com",
//				initiateURL:      "https://www.domain.com/oauth/initiate?provider=ixtendio1",
//				authorizeURL:     "https://www.domain.com/oauth/authorize/ixtendio1?state=123456&code=code123",
//				fetchUserDetails: true,
//			},
//			want: want{
//				initiateResponse:    `<a href="https://www.domain.com/oauth/authorize/ixtendio1?state=123456&amp;profile=true">Found</a>`,
//				authorizeResponse:   "initiateOAUTHMiddleware|authorizeOAUTHMiddleware|at=https://www.domain.com/oauth/authorize/ixtendio1:code123;sp=user123",
//				initiateStatusCode:  http.StatusFound,
//				authorizeStatusCode: 200,
//			},
//		},
//		{
//			name: "without fetching user details",
//			args: args{
//				providers:        []oauth.Provider{captureOAUTHProvider{name: "ixtendio"}},
//				websiteUrl:       "https://www.domain.com",
//				initiateURL:      "https://www.domain.com/oauth/initiate",
//				authorizeURL:     "https://www.domain.com/oauth/authorize/ixtendio?state=123456&code=code123",
//				fetchUserDetails: false,
//			},
//			want: want{
//				initiateResponse:    `<a href="https://www.domain.com/oauth/authorize/ixtendio?state=123456&amp;profile=false">Found</a>`,
//				authorizeResponse:   "initiateOAUTHMiddleware|authorizeOAUTHMiddleware|at=https://www.domain.com/oauth/authorize/ixtendio:code123",
//				initiateStatusCode:  http.StatusFound,
//				authorizeStatusCode: 200,
//			},
//		},
//		{
//			name: "using 2 providers but not specify it in the url",
//			args: args{
//				providers:        []oauth.Provider{captureOAUTHProvider{name: "ixtendio"}, captureOAUTHProvider{name: "ixtendio1"}},
//				websiteUrl:       "https://www.domain.com",
//				initiateURL:      "https://www.domain.com/oauth/initiate?provider=ixtendio2",
//				authorizeURL:     "https://www.domain.com/oauth/authorize/ixtendio1?state=123456&code=code123",
//				fetchUserDetails: true,
//			},
//			want: want{
//				initiateStatusCode: 500,
//			},
//		},
//		{
//			name: "specify an unsupported provider",
//			args: args{
//				providers:        []oauth.Provider{captureOAUTHProvider{name: "ixtendio"}, captureOAUTHProvider{name: "ixtendio1"}},
//				websiteUrl:       "https://www.domain.com",
//				initiateURL:      "https://www.domain.com/oauth/initiate",
//				authorizeURL:     "https://www.domain.com/oauth/authorize/ixtendio1?state=123456&code=code123",
//				fetchUserDetails: true,
//			},
//			want: want{
//				initiateStatusCode: 500,
//			},
//		},
//		{
//			name: "OAUTH provider returns error",
//			args: args{
//				providers:        []oauth.Provider{captureOAUTHProvider{name: "ixtendio"}},
//				websiteUrl:       "https://www.domain.com",
//				initiateURL:      "https://www.domain.com/oauth/initiate",
//				authorizeURL:     "https://www.domain.com/oauth/authorize/ixtendio?error=anerror",
//				fetchUserDetails: true,
//			},
//			want: want{
//				initiateResponse:    `<a href="https://www.domain.com/oauth/authorize/ixtendio?state=123456&amp;profile=true">Found</a>`,
//				authorizeResponse:   "initiateOAUTHMiddleware|authorizeOAUTHMiddleware|",
//				initiateStatusCode:  http.StatusFound,
//				authorizeStatusCode: 500,
//			},
//		},
//		{
//			name: "OAUTH provider returns a wrong state value",
//			args: args{
//				providers:        []oauth.Provider{captureOAUTHProvider{name: "ixtendio"}},
//				websiteUrl:       "https://www.domain.com",
//				initiateURL:      "https://www.domain.com/oauth/initiate",
//				authorizeURL:     "https://www.domain.com/oauth/authorize/ixtendio?state=1234568&code=code123",
//				fetchUserDetails: true,
//			},
//			want: want{
//				initiateResponse:    `<a href="https://www.domain.com/oauth/authorize/ixtendio?state=123456&amp;profile=true">Found</a>`,
//				authorizeResponse:   "initiateOAUTHMiddleware|authorizeOAUTHMiddleware|",
//				initiateStatusCode:  http.StatusFound,
//				authorizeStatusCode: 500,
//			},
//		},
//		{
//			name: "using wrong OAUTH provider name in the authorizeURL",
//			args: args{
//				providers:        []oauth.Provider{captureOAUTHProvider{name: "ixtendio"}},
//				websiteUrl:       "https://www.domain.com",
//				initiateURL:      "https://www.domain.com/oauth/initiate",
//				authorizeURL:     "https://www.domain.com/oauth/authorize/ixtendio1?state=123456&code=code123",
//				fetchUserDetails: true,
//			},
//			want: want{
//				initiateResponse:    `<a href="https://www.domain.com/oauth/authorize/ixtendio?state=123456&amp;profile=true">Found</a>`,
//				authorizeResponse:   "initiateOAUTHMiddleware|authorizeOAUTHMiddleware|",
//				initiateStatusCode:  http.StatusFound,
//				authorizeStatusCode: 500,
//			},
//		},
//	}
//	for _, tt := range tests {
//		sb.Reset()
//		t.Run(tt.name, func(t *testing.T) {
//			m, _ := NewMuxHandlerWithDefaultConfig()
//			m.HandleOAUTH2(oauth.Config{
//				WebsiteUrl:       tt.args.websiteUrl,
//				FetchUserDetails: tt.args.fetchUserDetails,
//				Providers:        tt.args.providers,
//				CacheConfig: oauth.CacheConfig{
//					Cache:             cache.NewInMemory(),
//					KeyExpirationTime: 60 * time.Second,
//				},
//			}, func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
//				at := oauth.GetAccessTokenFromContext(ctx)
//				sb.WriteString("at=" + at.AccessToken)
//				if tt.args.fetchUserDetails {
//					usr := auth.GetSecurityPrincipalFromContext(ctx)
//					sb.WriteString(";sp=" + usr.Identity())
//				}
//				return response.PlainTextHttpResponseOK(""), nil
//			}, []middleware.Middleware{func(handler handler.Handler) handler.Handler {
//				return func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
//					sb.WriteString("initiateOAUTHMiddleware|")
//					return handler(ctx, r)
//				}
//			}}, []middleware.Middleware{func(handler handler.Handler) handler.Handler {
//				return func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
//					sb.WriteString("authorizeOAUTHMiddleware|")
//					return handler(ctx, r)
//				}
//			}})
//
//			initiateReq, err := http.NewRequest(http.MethodGet, tt.args.initiateURL, nil)
//			if err != nil {
//				t.Errorf("HandleOAUTH2() failed to create initiate request: %v", err)
//			}
//			initResponseRecorder := httptest.NewRecorder()
//			m.ServeHTTP(initResponseRecorder, initiateReq)
//			if initResponseRecorder.Code != tt.want.initiateStatusCode {
//				t.Fatalf("HandleOAUTH2() got initiate responseCode: %v, want: %v", initResponseRecorder.Code, tt.want.initiateStatusCode)
//			}
//			responseBody := initResponseRecorder.Body.String()
//			if !strings.Contains(responseBody, tt.want.initiateResponse) {
//				t.Fatalf("HandleOAUTH2() got initiate response: '%v', want: '%v'", responseBody, tt.want.initiateResponse)
//			}
//
//			if initResponseRecorder.Code == http.StatusFound {
//				authorizeResponseRecorder := httptest.NewRecorder()
//				authorizeReq, err := http.NewRequest(http.MethodGet, tt.args.authorizeURL, nil)
//				if err != nil {
//					t.Fatalf("HandleOAUTH2() failed to create authorize request: %v", err)
//				}
//				m.ServeHTTP(authorizeResponseRecorder, authorizeReq)
//				if authorizeResponseRecorder.Code != tt.want.authorizeStatusCode {
//					t.Fatalf("HandleOAUTH2() got authorize responseCode: %v, want: %v", authorizeResponseRecorder.Code, tt.want.authorizeStatusCode)
//				}
//				if sb.String() != tt.want.authorizeResponse {
//					t.Fatalf("HandleOAUTH2() got authorize response: %v, want: %v", sb.String(), tt.want.authorizeResponse)
//				}
//			}
//		})
//	}
//}
//
//func TestMuxHandler_EnableDebugEndpoints(t *testing.T) {
//	tests := []struct {
//		name string
//	}{
//		{name: "register debug endpoints"},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			m, _ := NewMuxHandlerWithDefaultConfig()
//			m.EnableDebugEndpoints()
//
//			endpoints := []string{"/debug/pprof/", "/debug/pprof/allocs", "/debug/pprof/block",
//				"/debug/pprof/goroutine", "/debug/pprof/heap", "/debug/pprof/mutex",
//				"/debug/pprof/threadcreate", "/debug/pprof/cmdline", "/debug/pprof/profile",
//				"/debug/pprof/symbol", "/debug/pprof/trace", "/debug/vars"}
//			for _, ep := range endpoints {
//				req, _ := http.NewRequest(http.MethodGet, "https://www.domain.com"+ep, nil)
//				matchRequest, _ := m.router.MatchRequest(req)
//				if matchRequest == nil {
//					t.Errorf("EnableDebugEndpoints() the endpoint: %s was not found", ep)
//				}
//			}
//
//		})
//	}
//}
//
//func TestMuxHandler_GenerateUniqueId(t *testing.T) {
//	tests := []struct {
//		name string
//	}{
//		{name: "generate unique id"},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			genCount := 1000
//			idLen := 13
//			m := make(map[string]bool)
//			for i := 0; i < genCount; i++ {
//				id := generateUniqueId(idLen)
//				if len(id) != idLen {
//					t.Errorf("GenerateUniqueId() the unique id lenght is: %v but want %v", len(id), idLen)
//				}
//				m[id] = true
//			}
//			if len(m) != genCount {
//				t.Errorf("GenerateUniqueId() collison id detected: got: %v, want: %v", len(m), genCount)
//			}
//
//		})
//	}
//}
//
//func mustParseURL(rawURL string) *url.URL {
//	u, err := url.Parse(rawURL)
//	if err != nil {
//		log.Fatalf("Failed parsing the url: %s, err:%v", rawURL, err)
//	}
//	return u
//}
//
//func compareMiddlewares(m1 []middleware.Middleware, m2 []middleware.Middleware) error {
//	if len(m1) != len(m2) {
//		return fmt.Errorf("length => got: %v, want: %v", len(m1), len(m2))
//	}
//	for i := 0; i < len(m1); i++ {
//		m1FuncName := runtime.FuncForPC(reflect.ValueOf(m1[i]).Pointer()).Name()
//		m2FuncName := runtime.FuncForPC(reflect.ValueOf(m2[i]).Pointer()).Name()
//		if m1FuncName != m2FuncName {
//			return fmt.Errorf("istance => index: %v, got: %v, want: %v", i, m1FuncName, m2FuncName)
//		}
//	}
//	return nil
//}
