package router

import (
	"context"
	"fmt"
	"github.com/ixtendio/gow/internal/path"
	"github.com/ixtendio/gow/request"
	"github.com/ixtendio/gow/response"
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
	errLogFunc               func(err error)
	endpointMatcher          *matcher
}

func NewRouter(caseInsensitivePathMatch bool, errLogFunc func(err error)) *Router {
	if errLogFunc == nil {
		errLogFunc = defaultErrLogFunc
	}
	return &Router{
		caseInsensitivePathMatch: caseInsensitivePathMatch,
		errLogFunc:               errLogFunc,
		endpointMatcher:          newMatcher(),
	}
}

func (r *Router) Handle(method string, pathPattern string, handler Handler) {
	if err := r.endpointMatcher.addEndpoint(method, pathPattern, r.caseInsensitivePathMatch, handler); err != nil {
		r.errLogFunc(fmt.Errorf("failed registring the pathPattern: %s, err: %w", pathPattern, err))
		os.Exit(1)
	}
}

// ServeHTTP implements the http.Handler interface.
// It's the entry point for all http traffic
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	mc := path.ParseRequestURL(req.URL)
	handler := r.endpointMatcher.match(mc)
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
		RawRequest: req,
		UriVars:    mc.ExtractedUriVariables,
	}

	// Call the wrapped handler functions.
	resp, err := handler(ctx, &httpRequest)
	if err != nil {
		r.errLogFunc(fmt.Errorf("uncaught error, err: %w", err))
		resp = response.InternalServerErrorHttpResponse()
	}

	if err := resp.Write(w, &httpRequest); err != nil {
		r.errLogFunc(err)
	}
}
