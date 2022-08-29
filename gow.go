package gow

import (
	"context"
	"expvar"
	"fmt"
	"github.com/ixtendio/gow/request"
	"github.com/ixtendio/gow/response"
	"github.com/ixtendio/gow/router"
	"html/template"
	"log"
	"net/http"
	"net/http/pprof"
)

var defaultTemplateFunc = func(templatesPathPattern string) (*template.Template, error) {
	return template.New("").Funcs(template.FuncMap{
		"safe": func(s string) template.HTML { return template.HTML(s) }, //https://stackoverflow.com/questions/34348072/go-html-comments-are-not-rendered
	}).ParseGlob(templatesPathPattern)
}

// A Middleware is a function that receives a Handler and returns another Handler
type Middleware func(handler router.Handler) router.Handler

type TemplateConfig struct {
	//the dir path pattern that matches the Go templates. Default: "resources/templates/*.html"
	TemplatesPathPattern string
	//the dir path to the static resources (HTML pages, JS pages, IMAGES, etc). Default: "./resources/assets"
	AssetsDirPath string
	//the web path to server the static resources. Default: "assets"
	AssetsPath string
	//the Go templates. Default: template.HTML
	Template *template.Template
}

// A Config is a type used to pass the configurations to the Gow
type Config struct {
	//if the path match should be case-sensitive or not. Default false
	CaseInsensitivePathMatch bool
	//the application context path. Default: "/"
	ContextPath string
	//the TemplateConfig if the application supports static resources and templates. Default: nil
	TemplateConfig *TemplateConfig
	//a log function for critical errors. Default: defaultErrLogFunc
	ErrLogFunc func(err error)
}

func (c *Config) setDefaults() error {
	if c.ContextPath == "" {
		c.ContextPath = "/"
	}
	if c.ErrLogFunc == nil {
		c.ErrLogFunc = func(err error) {
			log.Printf("An error occured while handling the request, err: %v\n", err)
		}
	}
	if c.TemplateConfig != nil {
		if c.TemplateConfig.TemplatesPathPattern == "" {
			c.TemplateConfig.TemplatesPathPattern = "resources/templates/*.html"
		}
		if c.TemplateConfig.AssetsDirPath == "" {
			c.TemplateConfig.AssetsDirPath = "./resources/assets"
		}
		if c.TemplateConfig.AssetsPath == "" {
			c.TemplateConfig.AssetsPath = "assets"
		}
		if c.TemplateConfig.Template == nil {
			tmpl, err := defaultTemplateFunc(c.TemplateConfig.TemplatesPathPattern)
			if err != nil {
				return fmt.Errorf("failed parsing the templates, err: %w", err)
			}
			c.TemplateConfig.Template = tmpl
		}
	}
	return nil
}

type Gow struct {
	router            *router.Router
	commonMiddlewares []Middleware
	webConfig         *Config
}

// NewGow creates a new Gow instance
func NewGow(config *Config) (*Gow, error) {
	if err := config.setDefaults(); err != nil {
		return nil, err
	}
	r := router.NewRouter(config.CaseInsensitivePathMatch, config.ErrLogFunc)
	if config.TemplateConfig != nil {
		contextPath := config.ContextPath
		if contextPath == "/" {
			contextPath = ""
		}
		assetsPath := config.TemplateConfig.AssetsPath
		assetsDirPath := config.TemplateConfig.AssetsDirPath
		r.Handle(http.MethodGet, fmt.Sprintf("%s/%s/*", contextPath, assetsPath), router.Handler2Handler(http.StripPrefix(fmt.Sprintf("%s/%s/", contextPath, assetsPath), http.FileServer(http.Dir(assetsDirPath)))))
	}
	return &Gow{
		router:    r,
		webConfig: config,
	}, nil
}

// RegisterCommonMiddlewares allows registering common middlewares
func (g *Gow) RegisterCommonMiddlewares(middlewares ...Middleware) {
	for _, middleware := range middlewares {
		g.commonMiddlewares = append(g.commonMiddlewares, middleware)
	}
}

//// HandleAssetsResources add a handler for serving the static resources
//func (a *Gow) HandleAssetsResources() {
//	contextPath := a.contextPath
//	if a.templateConfig != nil {
//		assetsPath := a.assetsPath
//		assetsDirPath := a.assetsDirPath
//		a.mux.Handler(http.MethodGet, fmt.Sprintf("%s/%s/*", contextPath, assetsPath), http.StripPrefix(fmt.Sprintf("%s/%s/", contextPath, assetsPath), http.FileServer(http.Dir(assetsDirPath))))
//	}
//}

// HandleGet add a handler for handling a GET request
func (g *Gow) HandleGet(path string, handler router.Handler, middlewares ...Middleware) {
	g.HandleRequest(http.MethodGet, path, handler, middlewares...)
}

// HandlePost add a handler for handling a POST request
func (g *Gow) HandlePost(path string, handler router.Handler, middlewares ...Middleware) {
	g.HandleRequest(http.MethodPost, path, handler, middlewares...)
}

// HandlePut add a handler for handling a PUT request
func (g *Gow) HandlePut(path string, handler router.Handler, middlewares ...Middleware) {
	g.HandleRequest(http.MethodPut, path, handler, middlewares...)
}

// HandleDelete add a handler for handling a DELETE request
func (g *Gow) HandleDelete(path string, handler router.Handler, middlewares ...Middleware) {
	g.HandleRequest(http.MethodDelete, path, handler, middlewares...)
}

func (g *Gow) HandleRequest(method string, path string, handler router.Handler, middlewares ...Middleware) {
	handler = wrapMiddleware(wrapMiddleware(handler, middlewares...), g.commonMiddlewares...)
	var tmpl *template.Template
	if g.webConfig.TemplateConfig != nil {
		tmpl = g.webConfig.TemplateConfig.Template
	}
	//expose contextPath and template on request context
	handler = wrapMiddleware(handler, func(handler router.Handler) router.Handler {
		return func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
			ctx = context.WithValue(ctx, KeyValues, &CtxValues{
				ContextPath: g.webConfig.ContextPath,
				Template:    tmpl,
			})
			return handler(ctx, r)
		}
	})
	g.router.Handle(method, path, handler)
}

// EnableDebugEndpoints enable debug endpoints
func (g Gow) EnableDebugEndpoints() {
	// Register all the standard library debug endpoints.
	g.router.Handle(http.MethodGet, "/debug/pprof/", router.HandlerFunc2Handler(pprof.Index))
	g.router.Handle(http.MethodGet, "/debug/pprof/allocs", router.HandlerFunc2Handler(pprof.Index))
	g.router.Handle(http.MethodGet, "/debug/pprof/block", router.HandlerFunc2Handler(pprof.Index))
	g.router.Handle(http.MethodGet, "/debug/pprof/goroutine", router.HandlerFunc2Handler(pprof.Index))
	g.router.Handle(http.MethodGet, "/debug/pprof/heap", router.HandlerFunc2Handler(pprof.Index))
	g.router.Handle(http.MethodGet, "/debug/pprof/mutex", router.HandlerFunc2Handler(pprof.Index))
	g.router.Handle(http.MethodGet, "/debug/pprof/threadcreate", router.HandlerFunc2Handler(pprof.Index))
	g.router.Handle(http.MethodGet, "/debug/pprof/cmdline", router.HandlerFunc2Handler(pprof.Cmdline))
	g.router.Handle(http.MethodGet, "/debug/pprof/profile", router.HandlerFunc2Handler(pprof.Profile))
	g.router.Handle(http.MethodGet, "/debug/pprof/symbol", router.HandlerFunc2Handler(pprof.Symbol))
	g.router.Handle(http.MethodGet, "/debug/pprof/trace", router.HandlerFunc2Handler(pprof.Trace))
	g.router.Handle(http.MethodGet, "/debug/vars", router.Handler2Handler(expvar.Handler()))
}

func wrapMiddleware(handler router.Handler, middlewares ...Middleware) router.Handler {
	wrappedHandlers := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		middleware := middlewares[i]
		if middleware != nil {
			wrappedHandlers = middleware(wrappedHandlers)
		}
	}
	return wrappedHandlers
}
