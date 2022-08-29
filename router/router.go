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
	endpoints                []*endpoint
}

func NewRouter(caseInsensitivePathMatch bool, errLogFunc func(err error)) *Router {
	if errLogFunc == nil {
		errLogFunc = defaultErrLogFunc
	}
	return &Router{
		caseInsensitivePathMatch: caseInsensitivePathMatch,
		errLogFunc:               errLogFunc,
	}
}

func (r *Router) Handle(method string, apiPath string, handler Handler) {
	rootPath, err := path.Parse(apiPath, r.caseInsensitivePathMatch)
	if err != nil {
		r.errLogFunc(fmt.Errorf("failed parsing the path: %s, err: %w", apiPath, err))
		os.Exit(1)
	}
	r.endpoints = append(r.endpoints, &endpoint{
		method:   method,
		rootPath: rootPath,
		handler:  handler,
	})
}

// ServeHTTP implements the http.Handler interface.
// It's the entry point for all http traffic
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	mc := path.ParseRequestURL(req.URL)
	var e *endpoint
	if e == nil {

	}
	handler := e.handler
	// Pull the context from the request and use it as a separate parameter.
	ctx := req.Context()

	// Set the context with the required values to process the request.
	ctx = context.WithValue(ctx, KeyValues, &CtxValues{
		CorrelationId: fmt.Sprintf("%d:%d", time.Now().UnixNano(), rand.Int()),
		StartTime:     time.Now(),
	})

	httpRequest := request.HttpRequest{
		RawRequest: req,
		PathVars:   mc.ExtractedUriVariables,
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
