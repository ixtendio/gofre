package middleware

import (
	"context"
	"github.com/ixtendio/gofre/errors"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/request"
	"github.com/ixtendio/gofre/response"
	"testing"
)

func TestCompressResponse(t *testing.T) {
	type args struct {
		compressionLevel int
		handler          handler.Handler
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy flow",
			args: args{
				handler: func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
					return response.PlainTextHttpResponseOK("ok"), nil
				},
			},
			wantErr: false,
		},
		{
			name: "the handler error should be returned",
			args: args{
				handler: func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
					return nil, errors.ErrUnauthorized
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := CompressResponse(tt.args.compressionLevel)(tt.args.handler)(context.Background(), &request.HttpRequest{})
			if (err != nil) != tt.wantErr {
				t.Errorf("CompressResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if _, ok := resp.(*response.HttpCompressResponse); !ok {
					t.Errorf("CompressResponse() resp = %T, want *response.HttpCompressResponse", resp)
				}
			}
		})
	}
}
