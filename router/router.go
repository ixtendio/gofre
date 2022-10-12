package router

import (
	"context"
	"fmt"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/internal/path"
	"github.com/ixtendio/gofre/request"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var defaultErrLogFunc = func(err error) {
	log.Printf("%v", err)
}

type Router struct {
	caseInsensitivePathMatch bool
	endpointMatcher          *matcher
	errLogFunc               func(err error)
}

func NewRouter(caseInsensitivePathMatch bool, errLogFunc func(err error)) *Router {
	if errLogFunc == nil {
		errLogFunc = defaultErrLogFunc
	}
	return &Router{
		caseInsensitivePathMatch: caseInsensitivePathMatch,
		endpointMatcher:          newMatcher(),
		errLogFunc:               errLogFunc,
	}
}

func (r *Router) Handle(method string, pathPattern string, handler handler.Handler) *Router {
	if err := r.endpointMatcher.addEndpoint(method, pathPattern, r.caseInsensitivePathMatch, handler); err != nil {
		panic(fmt.Errorf("failed to register the path pattern: %s, err: %w", pathPattern, err))
	}
	return r
}

// ServeHTTP implements the http.Handler interface.
// It's the entry point for all http traffic
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	mc := path.ParseURL(req.URL)
	matchedHandler, capturedVars := r.endpointMatcher.match(req.Method, mc)
	if matchedHandler == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Pull the context from the request and use it as a separate parameter.
	ctx := req.Context()

	// Set the context with the required values to process the request.
	ctx = context.WithValue(ctx, KeyValues, &CtxValues{
		CorrelationId: fmt.Sprintf("%d:%d", time.Now().UnixNano(), rand.Int()),
		StartTime:     time.Now(),
	})

	httpRequest := request.HttpRequest{
		R:       req,
		UriVars: capturedVars,
	}

	// Call the wrapped handler functions.
	resp, err := matchedHandler(ctx, &httpRequest)
	if err != nil {
		r.errLogFunc(fmt.Errorf("uncaught error, err: %w", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := resp.Write(w, &httpRequest); err != nil {
		r.errLogFunc(err)
	}
}
