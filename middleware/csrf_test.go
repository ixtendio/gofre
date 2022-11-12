package middleware

import (
	"context"
	"github.com/ixtendio/gofre/cache"
	"github.com/ixtendio/gofre/errors"
	"github.com/ixtendio/gofre/request"
	"github.com/ixtendio/gofre/response"
	"net/http"
	"testing"
	"unicode"
)

func Test_generateNonce(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Run("generateNonce", func(t *testing.T) {
			nonce, err := generateNonce()
			if err != nil {
				t.Errorf("generateNonce() returned error: %v", err)
				return
			}
			for _, c := range nonce {
				if !unicode.IsDigit(c) && (c < 'A' || c > 'F') {
					t.Errorf("generateNonce() %s, is not HEXA value", nonce)
				}
			}
		})
	}
}

func TestCSRFPrevention(t *testing.T) {
	type args struct {
		nonceInCache string
		req          *http.Request
	}
	tests := []struct {
		name      string
		args      args
		wantError error
	}{
		{
			name:      "skip check",
			args:      args{req: &http.Request{Method: http.MethodGet}},
			wantError: nil,
		},
		{
			name:      "nonce check fails - header not present",
			args:      args{req: &http.Request{Method: http.MethodPost}},
			wantError: errors.ErrAccessDenied,
		},
		{
			name: "nonce check fails - wrong nonce in header",
			args: args{
				nonceInCache: "12345",
				req:          &http.Request{Method: http.MethodPost, Header: http.Header{CSRFRestNonceHeaderName: {"123"}}},
			},
			wantError: errors.ErrAccessDenied,
		},
		{
			name: "nonce check pass",
			args: args{
				nonceInCache: "12345",
				req:          &http.Request{Method: http.MethodPost, Header: http.Header{CSRFRestNonceHeaderName: {"12345"}}},
			},
			wantError: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memoryCache := cache.NewInMemory()
			if tt.args.nonceInCache != "" {
				memoryCache.Add(tt.args.nonceInCache, CSRFExpirationTime)
			}
			m := CSRFPrevention(memoryCache)
			_, err := m(func(ctx context.Context, r request.HttpRequest) (response.HttpResponse, error) {
				nonce := GetCSRFNonceFromContext(ctx)
				if nonce == "" {
					t.Fatalf("CSRFPrevention() nonce is empty")
				}
				return response.PlainTextHttpResponseOK("ok"), nil
			})(context.Background(), request.HttpRequest{R: tt.args.req})

			if err != tt.wantError {
				t.Fatalf("CSRFPrevention() error = %v, want %v", err, tt.wantError)
			}
		})
	}
}
