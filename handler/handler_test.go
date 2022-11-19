package handler

import (
	"context"
	"github.com/ixtendio/gofre/router/path"

	response2 "github.com/ixtendio/gofre/response"
	"net/http"
	"testing"
)

func TestHandler2Handler(t *testing.T) {
	type args struct {
		handler http.Handler
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "constructor",
			args: args{handler: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := Handler2Handler(tt.args.handler)(context.Background(), path.MatchingContext{})
			if err != nil {
				t.Fatalf("Handler2Handler() returned error: %v", err)
			}
			if _, ok := response.(*response2.HttpHandlerAdaptorResponse); !ok {
				t.Fatalf("Handler2Handler() expected instance of *response.HttpHandlerAdaptorResponse , got: %T", response)
			}
		})
	}
}

func TestHandlerFunc2Handler(t *testing.T) {
	type args struct {
		handlerFunc http.HandlerFunc
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "constructor",
			args: args{handlerFunc: func(writer http.ResponseWriter, request *http.Request) {

			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := Handler2Handler(tt.args.handlerFunc)(context.Background(), path.MatchingContext{})
			if err != nil {
				t.Fatalf("HandlerFunc2Handler() returned error: %v", err)
			}
			if _, ok := response.(*response2.HttpHandlerAdaptorResponse); !ok {
				t.Fatalf("HandlerFunc2Handler() expected instance of *response.HttpHandlerAdaptorResponse , got: %T", response)
			}
		})
	}
}
