package router

import (
	"context"
	"fmt"
	"github.com/ixtendio/gow/internal/path"
	"github.com/ixtendio/gow/request"
	"log"
	"math/rand"
	"net/http"
	"os"
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

func (r *Router) Handle(method string, pathPattern string, handler Handler) *Router {
	if err := r.endpointMatcher.addEndpoint(method, pathPattern, r.caseInsensitivePathMatch, handler); err != nil {
		r.errLogFunc(fmt.Errorf("failed registring the pathPattern: %s, err: %writer", pathPattern, err))
		os.Exit(1)
	}
	return r
}

// ServeHTTP implements the http.Handler interface.
// It's the entry point for all http traffic
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	mc := path.ParseURL(req.URL)
	handler, capturedVars := r.endpointMatcher.match(req.Method, mc)
	if handler == nil {
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
	resp, err := handler(ctx, &httpRequest)
	if err != nil {
		r.errLogFunc(fmt.Errorf("uncaught error, err: %w", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := resp.Write(w, &httpRequest); err != nil {
		r.errLogFunc(err)
	}
}
