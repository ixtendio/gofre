package handler

import (
	"context"
	"github.com/ixtendio/gofre/response"
	"github.com/ixtendio/gofre/router/path"
	"net/http"
)

// A Handler is a function that process a path.MatchingContext and returns a response.HttpResponse or an error
type Handler func(ctx context.Context, mc path.MatchingContext) (response.HttpResponse, error)

// The HandlerFunc2Handler adapts a GO http.HandlerFunc to a Handler
func HandlerFunc2Handler(handlerFunc http.HandlerFunc) Handler {
	return func(ctx context.Context, mc path.MatchingContext) (response.HttpResponse, error) {
		return response.HandlerFuncAdaptor(handlerFunc), nil
	}
}

// The Handler2Handler adapts a GO http.Handler to a Handler
func Handler2Handler(handler http.Handler) Handler {
	return func(ctx context.Context, mc path.MatchingContext) (response.HttpResponse, error) {
		return response.HandlerAdaptor(handler), nil
	}
}
