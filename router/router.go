package router

import (
	"fmt"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/router/path"
	"log"
	"net/http"
	"strings"
	"sync"
)

var urlPathSegmentsPool = sync.Pool{
	New: func() interface{} {
		arr := make([]path.UrlSegment, path.MaxPathSegments)
		return &arr
	},
}

var defaultErrLogFunc = func(err error) {
	log.Printf("%v", err)
}

type Router struct {
	caseInsensitivePathMatch bool
	endpointMatchers         map[string]*path.Matcher
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
		endpointMatchers:         make(map[string]*path.Matcher, 9),
		errLogFunc:               errLogFunc,
	}
}

// Handle register a new handler or panic if the handler can not be registered
// This method returns the router so that chain handler registration to be possible
func (r *Router) Handle(httpMethod string, pathPattern string, handler handler.Handler) *Router {
	pattern, err := path.ParsePattern(pathPattern, r.caseInsensitivePathMatch)
	if err != nil {
		panic(fmt.Sprintf("failed to parse match pattern: %s:%s, err: %v", httpMethod, pathPattern, err))
	}
	pattern.Attachment = handler
	httpMethod = strings.ToUpper(httpMethod)
	matcher := r.endpointMatchers[httpMethod]
	if matcher == nil {
		matcher = path.NewMatcher(r.caseInsensitivePathMatch)
		r.endpointMatchers[httpMethod] = matcher
	}
	if err := matcher.AddPattern(pattern); err != nil {
		panic(fmt.Sprintf("failed to register match pattern: %s:%s, err: %v", httpMethod, pathPattern, err))
	}
	return r
}

// ServeHTTP implements the http.Handler interface.
// It's the entry point for all http traffic
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	httpMethod := req.Method
	urlPath := req.URL.Path
	urlSegmentsPtr := urlPathSegmentsPool.Get().(*[]path.UrlSegment)
	defer func() {
		urlSegments := *urlSegmentsPtr
		for i := 0; i < path.MaxPathSegments; i++ {
			s := &urlSegments[i]
			s.Reset()
		}
		urlPathSegmentsPool.Put(urlSegmentsPtr)
	}()
	mc := path.MatchingContext{R: req, PathSegments: *urlSegmentsPtr}
	path.ParseURLPath(req.URL, &mc)
	httpMethod = strings.ToUpper(httpMethod)
	matcher := r.endpointMatchers[httpMethod]
	if matcher == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	pattern := matcher.Match(urlPath, &mc)
	if pattern == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	matchedHandler := pattern.Attachment.(handler.Handler)

	// Call the wrapped handler functions.
	resp, err := matchedHandler(req.Context(), mc)
	if err != nil {
		r.errLogFunc(fmt.Errorf("uncaught error in GoFre framework, err: %w", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := resp.Write(w, mc); err != nil {
		r.errLogFunc(err)
	}
}
