package middleware

import (
	"context"
	"github.com/ixtendio/gofre/errors"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/request"
	"github.com/ixtendio/gofre/response"
	"reflect"
	"strings"
	"testing"
)

func TestPanicMiddleware(t *testing.T) {
	type want struct {
		resp         response.HttpResponse
		errMsgPrefix string
	}
	tests := []struct {
		name    string
		handler handler.Handler
		want
	}{
		{
			name: "error are returns",
			handler: func(ctx context.Context, r request.HttpRequest) (response.HttpResponse, error) {
				return nil, errors.ErrWrongCredentials
			},
			want: want{
				resp:         nil,
				errMsgPrefix: errors.ErrWrongCredentials.Error(),
			},
		},
		{
			name: "panic is handled and error returned",
			handler: func(ctx context.Context, r request.HttpRequest) (response.HttpResponse, error) {
				panic("a panic message")
			},
			want: want{
				resp:         nil,
				errMsgPrefix: "recover from panic",
			},
		},
		{
			name: "error is nil",
			handler: func(ctx context.Context, r request.HttpRequest) (response.HttpResponse, error) {
				return response.HtmlHttpResponseOK("ok"), nil
			},
			want: want{
				resp:         response.HtmlHttpResponseOK("ok"),
				errMsgPrefix: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := PanicRecover()
			resp, err := m(tt.handler)(context.Background(), request.HttpRequest{})
			if err != nil {
				if strings.Index(err.Error(), tt.want.errMsgPrefix) != 0 {
					t.Errorf("PanicRecover() = %v, want to start with: %v", err.Error(), tt.want.errMsgPrefix)
				}
			} else if !reflect.DeepEqual(tt.want.resp, resp) {
				t.Errorf("PanicRecover() = %v, want %v", resp, tt.want.resp)
			}
		})
	}
}
