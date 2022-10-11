package middleware

import (
	"context"
	goerrors "errors"
	"github.com/ixtendio/gofre/errors"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/request"
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
			name: "ErrDenied => StatusForbidden",
			args: args{
				handler: func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
					return nil, errors.ErrDenied
				},
			},
			want: response.HtmlHttpResponse(http.StatusForbidden, errors.ErrDenied.Error()),
		},
		{
			name: "ErrWrongCredentials => StatusForbidden",
			args: args{
				handler: func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
					return nil, errors.ErrWrongCredentials
				},
			},
			want: response.HtmlHttpResponse(http.StatusForbidden, errors.ErrWrongCredentials.Error()),
		},
		{
			name: "ErrUnauthorized => StatusUnauthorized",
			args: args{
				handler: func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
					return nil, errors.ErrUnauthorized
				},
			},
			want: response.HtmlHttpResponse(http.StatusUnauthorized, errors.ErrUnauthorized.Error()),
		},
		{
			name: "ErrObjectNotFound => StatusNotFound",
			args: args{
				handler: func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
					return nil, errors.NewErrObjectNotFoundWithMessage("object not found")
				},
			},
			want: response.HtmlHttpResponse(http.StatusNotFound, "object not found"),
		},
		{
			name: "ErrInvalidRequest => StatusBadRequest",
			args: args{
				handler: func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
					return nil, errors.NewErrInvalidRequestWithMessage("invalid request")
				},
			},
			want: response.HtmlHttpResponse(http.StatusBadRequest, "invalid request"),
		},
		{
			name: "custom error => StatusInternalServerError",
			args: args{
				handler: func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
					return nil, goerrors.New("custom error")
				},
			},
			want: response.HtmlHttpResponse(http.StatusInternalServerError, "custom error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := ErrResponse(responseSupplier)(tt.args.handler)(context.Background(), nil)
			if err != nil {
				t.Errorf("ErrResponse() returned error: %v", err)
			}
			if !reflect.DeepEqual(resp, tt.want) {
				t.Errorf("ErrResponse() = %v, want %v", resp, tt.want)
			}
		})
	}
}
