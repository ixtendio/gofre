package gofre

import (
	"context"
	"expvar"
	"fmt"
	"github.com/ixtendio/gofre/auth"
	"github.com/ixtendio/gofre/auth/oauth"
	"github.com/ixtendio/gofre/errors"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/middleware"
	"github.com/ixtendio/gofre/request"
	"github.com/ixtendio/gofre/response"
	"github.com/ixtendio/gofre/router"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/http/pprof"
	"unsafe"
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
	AssetsMappingPath string
	//the Go templates. Default: template.HTML
	Template response.ExecutableTemplate
}

func (c *ResourcesConfig) setDefaults() error {
	if c.TemplatesPathPattern == "" {
		c.TemplatesPathPattern = "resources/templates/*.html"
	}
	if c.AssetsDirPath == "" {
		c.AssetsDirPath = "./resources/assets"
	}
	if c.AssetsMappingPath == "" {
		c.AssetsMappingPath = "assets"
	}
	if c.Template == nil {
		tmpl, err := defaultTemplateFunc(c.TemplatesPathPattern)
		if err != nil {
			return fmt.Errorf("failed parsing the templates, err: %w", err)
		}
		c.Template = tmpl
	}
	return nil
}

// A Config is a type used to pass the configuration to the MuxHandler
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
		if err := c.ResourcesConfig.setDefaults(); err != nil {
			return err
		}
	}
	return nil
}

type MuxHandler struct {
	pathPrefix        string
	router            *router.Router
	commonMiddlewares []middleware.Middleware
	webConfig         *Config
}

// NewMuxHandlerWithDefaultConfig returns a new MuxHandler using default configuration
func NewMuxHandlerWithDefaultConfig() (*MuxHandler, error) {
	return NewMuxHandler(&Config{})
}

// NewMuxHandlerWithDefaultConfigAndTemplateSupport returns a new MuxHandler using default configuration, static resources and HTML templating support
func NewMuxHandlerWithDefaultConfigAndTemplateSupport() (*MuxHandler, error) {
	return NewMuxHandler(&Config{ResourcesConfig: NewDefaultResourcesConfig()})
}

// NewDefaultResourcesConfig returns a new ResourcesConfig with default configs
func NewDefaultResourcesConfig() *ResourcesConfig {
	rc := &ResourcesConfig{}
	if err := rc.setDefaults(); err != nil {
		log.Fatalf("failed to set the default values for ResourcesConfig, err: %v", err)
	}
	return rc
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
		assetsPath := config.ResourcesConfig.AssetsMappingPath
		assetsDirPath := config.ResourcesConfig.AssetsDirPath
		r.Handle(http.MethodGet, fmt.Sprintf("%s/%s/*", contextPath, assetsPath), handler.Handler2Handler(http.StripPrefix(fmt.Sprintf("%s/%s/", contextPath, assetsPath), http.FileServer(http.Dir(assetsDirPath)))))
	}
	return &MuxHandler{
		router:    r,
		webConfig: config,
	}, nil
}

// Config returns a Config copy
func (m *MuxHandler) Config() Config {
	return *m.webConfig
}

// ExecutableTemplate returns the response.ExecutableTemplate from ResourcesConfig or nil
func (m *MuxHandler) ExecutableTemplate() response.ExecutableTemplate {
	if m.webConfig.ResourcesConfig == nil {
		return nil
	}
	return m.webConfig.ResourcesConfig.Template
}

// RouteWithPathPrefix creates a new MuxHandler with a custom path prefix.
// The new mux handler will inherit the common middlewares from the parent, and the new common middlewares added
// to it will not be reflected to the parent.
// The new path prefix will equal with the parent path prefix + the new path prefix
func (m *MuxHandler) RouteWithPathPrefix(pathPrefix string) *MuxHandler {
	if len(pathPrefix) == 0 || pathPrefix == m.pathPrefix {
		return m
	}
	return &MuxHandler{
		pathPrefix:        m.resolvePath(pathPrefix),
		router:            m.router,
		commonMiddlewares: m.commonMiddlewares,
		webConfig:         m.webConfig,
	}
}

// CommonMiddlewares registers middlewares that will be applied for all handlers
func (m *MuxHandler) CommonMiddlewares(middlewares ...middleware.Middleware) {
	m.commonMiddlewares = append(m.commonMiddlewares, middlewares...)
}

// HandleOAUTH2 registers the necessary handlers to initiate and complete the OAUTH2 flow
//
// this method registers two endpoints:
// 1. GET: /oauth/initiate - initiate the OAUTH2 flow using a provider. If multiple providers are passed in the oauth.Config, then the parameter `provider` should be specified in the query string (example: /oauth/initiate?provider=github)
// 2. GET: /oauth/authorize/{provider} - (the redirect URI) exchange the authorization code for a JWT. The provider value is the name of the OAUTH2 provider (example: /oauth/authorize/github )
//
// If the OAUTH2 flow successfully completes, then the oauth.AccessToken will be passed to context.Context
// to extract it, you have to use the method oauth.GetAccessTokenFromContext(context.Context)
func (m *MuxHandler) HandleOAUTH2(oauthConfig oauth.Config, handler handler.Handler, initiateMiddlewares []middleware.Middleware, authorizeMiddlewares []middleware.Middleware) {
	m.HandleOAUTH2WithCustomPaths("/oauth/initiate", "/oauth/authorize", oauthConfig, handler, initiateMiddlewares, authorizeMiddlewares)
}

// HandleOAUTH2WithCustomPaths registers the necessary handlers to initiate and complete the OAUTH2 flow using custom paths
func (m *MuxHandler) HandleOAUTH2WithCustomPaths(initiatePath string,
	authorizeBasePath string,
	oauthConfig oauth.Config,
	handler handler.Handler,
	initiateMiddlewares []middleware.Middleware,
	authorizeMiddlewares []middleware.Middleware) {
	cache := oauthConfig.CacheConfig.Cache
	// initiate OAUTH flow handler
	authorizationFlowBasePath := authorizeBasePath
	m.HandleGet(initiatePath, func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		var provider oauth.Provider
		if len(oauthConfig.Providers) == 1 {
			provider = oauthConfig.Providers[0]
		} else {
			providerName := r.R.FormValue("provider")
			if providerName == "" {
				return nil, errors.NewBadRequestWithMessage("oauth provider not specified")
			}
			provider = oauthConfig.GetProviderByName(providerName)
		}
		if provider == nil {
			return nil, errors.NewBadRequestWithMessage("oauth provider not supported")
		}
		redirectUrl := oauthConfig.WebsiteUrl + m.resolvePath(authorizationFlowBasePath) + "/" + provider.Name()
		state := uniqueIdFunc(12)
		if cache != nil {
			if err := cache.Add(state, oauthConfig.CacheConfig.KeyExpirationTime); err != nil {
				return nil, fmt.Errorf("failed to save the OAUTH2 state random value, err: %w", err)
			}
		}

		return response.RedirectHttpResponse(provider.InitiateUrl(redirectUrl, state, oauthConfig.FetchUserDetails)), nil
	}, initiateMiddlewares...)

	// authorize OAUTH flow handler
	m.HandleGet(authorizationFlowBasePath+"/{providerName}", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		providerName := r.UriVars["providerName"]
		provider := oauthConfig.GetProviderByName(providerName)
		if provider == nil {
			return nil, errors.NewBadRequestWithMessage("oauth provider not supported")
		}

		redirectUrl := oauthConfig.WebsiteUrl + m.resolvePath(authorizationFlowBasePath) + "/" + provider.Name()
		errCode := r.R.FormValue("error")
		if errCode != "" {
			return nil, errors.ErrUnauthorizedRequest
		}

		state := r.R.FormValue("state")
		if cache != nil && !cache.Contains(state) {
			return nil, errors.ErrUnauthorizedRequest

		}
		code := r.R.FormValue("code")
		accessToken, err := provider.FetchAccessToken(ctx, redirectUrl, code)
		if err != nil {
			return nil, err
		}
		ctx = context.WithValue(ctx, oauth.AccessTokenCtxKey, accessToken)

		if oauthConfig.FetchUserDetails {
			user, err := provider.FetchAuthenticatedUser(ctx, accessToken)
			if err != nil {
				return nil, err
			}
			ctx = context.WithValue(ctx, auth.SecurityPrincipalCtxKey, &user)
		}

		return handler(ctx, r)
	}, authorizeMiddlewares...)
}

// HandleGet registers a handler with custom middlewares for GET requests
func (m *MuxHandler) HandleGet(path string, handler handler.Handler, middlewares ...middleware.Middleware) {
	m.HandleRequest(http.MethodGet, path, handler, middlewares...)
}

// HandlePost registers a handler with custom middlewares for POST requests
func (m *MuxHandler) HandlePost(path string, handler handler.Handler, middlewares ...middleware.Middleware) {
	m.HandleRequest(http.MethodPost, path, handler, middlewares...)
}

// HandlePut registers a handler with custom middlewares for PUT requests
func (m *MuxHandler) HandlePut(path string, handler handler.Handler, middlewares ...middleware.Middleware) {
	m.HandleRequest(http.MethodPut, path, handler, middlewares...)
}

// HandlePatch registers a handler with custom middlewares for PATCH requests
func (m *MuxHandler) HandlePatch(path string, handler handler.Handler, middlewares ...middleware.Middleware) {
	m.HandleRequest(http.MethodPatch, path, handler, middlewares...)
}

// HandleDelete registers a handler with custom middlewares for DELETE requests
func (m *MuxHandler) HandleDelete(path string, handler handler.Handler, middlewares ...middleware.Middleware) {
	m.HandleRequest(http.MethodDelete, path, handler, middlewares...)
}

// HandleRequest registers a handler with custom middlewares for the specified HTTP method
func (m *MuxHandler) HandleRequest(httpMethod string, path string, h handler.Handler, middlewares ...middleware.Middleware) {
	h = wrapMiddleware(wrapMiddleware(h, middlewares...), m.commonMiddlewares...)
	m.router.Handle(httpMethod, m.resolvePath(path), h)
}

func (m *MuxHandler) resolvePath(path string) string {
	pathPrefix := m.pathPrefix
	if len(pathPrefix) == 0 {
		return path
	}
	if len(path) == 0 {
		return pathPrefix
	}

	if pathPrefix[len(pathPrefix)-1] == '/' && path[0] == '/' {
		return pathPrefix + path[1:]
	} else if pathPrefix[len(pathPrefix)-1] != '/' && path[0] != '/' {
		return pathPrefix + "/" + path
	} else {
		return pathPrefix + path
	}
}

// EnableDebugEndpoints enable debug endpoints
func (m MuxHandler) EnableDebugEndpoints() {
	// Register all the standard library debug endpoints.
	m.router.Handle(http.MethodGet, m.resolvePath("/debug/pprof/"), handler.HandlerFunc2Handler(pprof.Index))
	m.router.Handle(http.MethodGet, m.resolvePath("/debug/pprof/allocs"), handler.HandlerFunc2Handler(pprof.Index))
	m.router.Handle(http.MethodGet, m.resolvePath("/debug/pprof/block"), handler.HandlerFunc2Handler(pprof.Index))
	m.router.Handle(http.MethodGet, m.resolvePath("/debug/pprof/goroutine"), handler.HandlerFunc2Handler(pprof.Index))
	m.router.Handle(http.MethodGet, m.resolvePath("/debug/pprof/heap"), handler.HandlerFunc2Handler(pprof.Index))
	m.router.Handle(http.MethodGet, m.resolvePath("/debug/pprof/mutex"), handler.HandlerFunc2Handler(pprof.Index))
	m.router.Handle(http.MethodGet, m.resolvePath("/debug/pprof/threadcreate"), handler.HandlerFunc2Handler(pprof.Index))
	m.router.Handle(http.MethodGet, m.resolvePath("/debug/pprof/cmdline"), handler.HandlerFunc2Handler(pprof.Cmdline))
	m.router.Handle(http.MethodGet, m.resolvePath("/debug/pprof/profile"), handler.HandlerFunc2Handler(pprof.Profile))
	m.router.Handle(http.MethodGet, m.resolvePath("/debug/pprof/symbol"), handler.HandlerFunc2Handler(pprof.Symbol))
	m.router.Handle(http.MethodGet, m.resolvePath("/debug/pprof/trace"), handler.HandlerFunc2Handler(pprof.Trace))
	m.router.Handle(http.MethodGet, m.resolvePath("/debug/vars"), handler.Handler2Handler(expvar.Handler()))
}

func (m *MuxHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m.router.ServeHTTP(w, req)
}

func wrapMiddleware(handler handler.Handler, middlewares ...middleware.Middleware) handler.Handler {
	if len(middlewares) == 0 {
		return handler
	}
	wrappedHandlers := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		mid := middlewares[i]
		if mid != nil {
			wrappedHandlers = mid(wrappedHandlers)
		}
	}
	return wrappedHandlers
}

const randLetters = "abcdefghijklmnopqrstuvwxyz1234567890"

var uniqueIdFunc = generateUniqueId

func generateUniqueId(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = randLetters[rand.Int63()%int64(len(randLetters))]
	}
	return *(*string)(unsafe.Pointer(&b))
}
