package middleware

import (
	"context"
	goerrors "errors"
	"github.com/ixtendio/gofre/errors"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/router/path"

	"github.com/ixtendio/gofre/response"
	"net/http"
	"reflect"
	"testing"
)

func TestErrResponse(t *testing.T) {
	type args struct {
		handler handler.Handler
	}
	responseSupplier := func(statusCode int, err error) response.HttpResponse {
		return response.HtmlHttpResponse(statusCode, err.Error())
	}
	tests := []struct {
		name string
		args args
		want response.HttpResponse
	}{
		{
			name: "ErrAccessDenied => StatusForbidden",
			args: args{
				handler: func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
					return nil, errors.ErrAccessDenied
				},
			},
			want: response.HtmlHttpResponse(http.StatusForbidden, errors.ErrAccessDenied.Error()),
		},
		{
			name: "ErrWrongCredentials => StatusForbidden",
			args: args{
				handler: func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
					return nil, errors.ErrWrongCredentials
				},
			},
			want: response.HtmlHttpResponse(http.StatusForbidden, errors.ErrWrongCredentials.Error()),
		},
		{
			name: "ErrUnauthorizedRequest => StatusUnauthorized",
			args: args{
				handler: func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
					return nil, errors.ErrUnauthorizedRequest
				},
			},
			want: response.HtmlHttpResponse(http.StatusUnauthorized, errors.ErrUnauthorizedRequest.Error()),
		},
		{
			name: "ErrObjectNotFound => StatusNotFound",
			args: args{
				handler: func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
					return nil, errors.NewObjectNotFoundWithMessage("object not found")
				},
			},
			want: response.HtmlHttpResponse(http.StatusNotFound, "object not found"),
		},
		{
			name: "ErrBadRequest => StatusBadRequest",
			args: args{
				handler: func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
					return nil, errors.NewBadRequestWithMessage("invalid request")
				},
			},
			want: response.HtmlHttpResponse(http.StatusBadRequest, "invalid request"),
		},
		{
			name: "custom error => StatusInternalServerError",
			args: args{
				handler: func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
					return nil, goerrors.New("custom error")
				},
			},
			want: response.HtmlHttpResponse(http.StatusInternalServerError, "custom error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := ErrResponse(responseSupplier)(tt.args.handler)(context.Background(), path.MatchingContext{})
			if err != nil {
				t.Fatalf("ErrResponse() returned error: %v", err)
			}
			if !reflect.DeepEqual(resp, tt.want) {
				t.Fatalf("ErrResponse() = %v, want %v", resp, tt.want)
			}
		})
	}
}

func TestErrJsonResponse(t *testing.T) {
	tests := []struct {
		name string
		want response.HttpResponse
	}{
		{
			name: "check json",
			want: response.JsonHttpResponse(http.StatusUnauthorized, map[string]string{
				"error": "unauthorized request",
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := ErrJsonResponse()(func(ctx context.Context, r path.MatchingContext) (response.HttpResponse, error) {
				return nil, errors.ErrUnauthorizedRequest
			})(context.Background(), path.MatchingContext{})
			if err != nil {
				t.Fatalf("ErrJsonResponse() returned error: %v", err)
			}
			if !reflect.DeepEqual(resp, tt.want) {
				t.Fatalf("ErrResponse() = %v, want %v", resp, tt.want)
			}
		})
	}
}
