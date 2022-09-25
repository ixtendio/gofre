package gow

import (
	"context"
	"expvar"
	"fmt"
	"github.com/ixtendio/gow/middleware"
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

type ResourcesConfig struct {
	//the dir path pattern that matches the Go templates. Default: "resources/templates/*.html"
	TemplatesPathPattern string
	//the dir path to the static resources (HTML pages, JS pages, Images, etc). Default: "./resources/assets"
	AssetsDirPath string
	//the web path to server the static resources. Default: "assets"
	AssetsPath string
	//the Go templates. Default: template.HTML
	Template *template.Template
}

// A Config is a type used to pass the configurations to the MuxHandler
type Config struct {
	//if the path match should be case-sensitive or not. Default false
	CaseInsensitivePathMatch bool
	//the application context path. Default: "/"
	ContextPath string
	//the ResourcesConfig if the application supports static resources and templates. Default: nil
	ResourcesConfig *ResourcesConfig
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
	if c.ResourcesConfig != nil {
		if c.ResourcesConfig.TemplatesPathPattern == "" {
			c.ResourcesConfig.TemplatesPathPattern = "resources/templates/*.html"
		}
		if c.ResourcesConfig.AssetsDirPath == "" {
			c.ResourcesConfig.AssetsDirPath = "./resources/assets"
		}
		if c.ResourcesConfig.AssetsPath == "" {
			c.ResourcesConfig.AssetsPath = "assets"
		}
		if c.ResourcesConfig.Template == nil {
			tmpl, err := defaultTemplateFunc(c.ResourcesConfig.TemplatesPathPattern)
			if err != nil {
				return fmt.Errorf("failed parsing the templates, err: %w", err)
			}
			c.ResourcesConfig.Template = tmpl
		}
	}
	return nil
}

type MuxHandler struct {
	router            *router.Router
	commonMiddlewares []middleware.Middleware
	webConfig         *Config
}

// NewMuxHandler creates a new MuxHandler instance
func NewMuxHandler(config *Config) (*MuxHandler, error) {
	if err := config.setDefaults(); err != nil {
		return nil, err
	}
	r := router.NewRouter(config.CaseInsensitivePathMatch, config.ErrLogFunc)
	if config.ResourcesConfig != nil {
		contextPath := config.ContextPath
		if contextPath == "/" {
			contextPath = ""
		}
		assetsPath := config.ResourcesConfig.AssetsPath
		assetsDirPath := config.ResourcesConfig.AssetsDirPath
		r.Handle(http.MethodGet, fmt.Sprintf("%s/%s/*", contextPath, assetsPath), router.Handler2Handler(http.StripPrefix(fmt.Sprintf("%s/%s/", contextPath, assetsPath), http.FileServer(http.Dir(assetsDirPath)))))
	}
	return &MuxHandler{
		router:    r,
		webConfig: config,
	}, nil
}

// RegisterCommonMiddlewares allows registering common middlewares
func (g *MuxHandler) RegisterCommonMiddlewares(middlewares ...middleware.Middleware) {
	for _, mid := range middlewares {
		g.commonMiddlewares = append(g.commonMiddlewares, mid)
	}
}

// HandleGet add a handler for handling a GET request
func (g *MuxHandler) HandleGet(path string, handler router.Handler, middlewares ...middleware.Middleware) {
	g.HandleRequest(http.MethodGet, path, handler, middlewares...)
}

// HandlePost add a handler for handling a POST request
func (g *MuxHandler) HandlePost(path string, handler router.Handler, middlewares ...middleware.Middleware) {
	g.HandleRequest(http.MethodPost, path, handler, middlewares...)
}

// HandlePut add a handler for handling a PUT request
func (g *MuxHandler) HandlePut(path string, handler router.Handler, middlewares ...middleware.Middleware) {
	g.HandleRequest(http.MethodPut, path, handler, middlewares...)
}

// HandleDelete add a handler for handling a DELETE request
func (g *MuxHandler) HandleDelete(path string, handler router.Handler, middlewares ...middleware.Middleware) {
	g.HandleRequest(http.MethodDelete, path, handler, middlewares...)
}

func (g *MuxHandler) HandleRequest(method string, path string, handler router.Handler, middlewares ...middleware.Middleware) {
	handler = wrapMiddleware(wrapMiddleware(handler, middlewares...), g.commonMiddlewares...)
	var tmpl *template.Template
	if g.webConfig.ResourcesConfig != nil {
		tmpl = g.webConfig.ResourcesConfig.Template
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
func (g MuxHandler) EnableDebugEndpoints() {
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

func wrapMiddleware(handler router.Handler, middlewares ...middleware.Middleware) router.Handler {
	wrappedHandlers := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		mid := middlewares[i]
		if mid != nil {
			wrappedHandlers = mid(wrappedHandlers)
		}
	}
	return wrappedHandlers
}

func (g *MuxHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	g.router.ServeHTTP(w, req)
}
