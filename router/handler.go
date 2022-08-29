package router

import (
	"context"
	"github.com/ixtendio/gow/request"
	"github.com/ixtendio/gow/response"
	"net/http"
)

// A Handler is a type that handles a http request
type Handler func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error)

// HandlerFunc2Handler wrap an http.HandlerFunc to an Handler
func HandlerFunc2Handler(handlerFunc http.HandlerFunc) Handler {
	return func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		return response.HandlerFuncAdaptor(handlerFunc), nil
	}
}

// Handler2Handler wrap an http.Handler to an Handler
func Handler2Handler(handler http.Handler) Handler {
	return func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		return response.HandlerAdaptor(handler), nil
	}
}
