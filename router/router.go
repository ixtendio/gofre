package router

import (
	"fmt"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/internal/path"
	"github.com/ixtendio/gofre/request"
	"log"
	"net/http"
	"net/url"
)

var defaultErrLogFunc = func(err error) {
	log.Printf("%v", err)
}

type Router struct {
	caseInsensitivePathMatch bool
	endpointMatcher          *matcher
	errLogFunc               func(err error)
}

func NewRouterWithDefaultConfig() *Router {
	return NewRouter(false, defaultErrLogFunc)
}
func NewRouter(caseInsensitivePathMatch bool, errLogFunc func(err error)) *Router {
	if errLogFunc == nil {
		errLogFunc = defaultErrLogFunc
	}
	return &Router{
		caseInsensitivePathMatch: caseInsensitivePathMatch,
		endpointMatcher:          newMatcheri(),
		errLogFunc:               errLogFunc,
	}
}

// Handle register a new handler or panic if the handler can not be registered
// This method returns the router so that chain handler registration to be possible
func (r *Router) Handle(method string, pathPattern string, handler handler.Handler) *Router {
	if err := r.endpointMatcher.addEndpoint(method, pathPattern, r.caseInsensitivePathMatch, handler); err != nil {
		panic(fmt.Sprintf("failed to register the path pattern: %s, err: %v", pathPattern, err))
	}
	return r
}

// MatchRequest returns the first handler that matches the request, together with the path variables if exists
func (r *Router) MatchRequest(req *http.Request) (handler.Handler, map[string]string) {
	return r.Match(req.Method, req.URL)
}

// Match returns the first handler that matches the http method and the url, together with the path variables if exists
func (r *Router) Match(httpMethod string, url *url.URL) (handler.Handler, map[string]string) {
	mc := path.ParseURLPath(url)
	return r.endpointMatcher.match(httpMethod, mc)
}

// ServeHTTP implements the http.Handler interface.
// It's the entry point for all http traffic
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	matchedHandler, capturedVars := r.MatchRequest(req)
	if matchedHandler == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	httpRequest := request.NewHttpRequestWithPathVars(req, capturedVars)
	// Call the wrapped handler functions.
	resp, err := matchedHandler(req.Context(), httpRequest)
	if err != nil {
		r.errLogFunc(fmt.Errorf("uncaught error in GoFre framework, err: %w", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := resp.Write(w, httpRequest); err != nil {
		r.errLogFunc(err)
	}
}
