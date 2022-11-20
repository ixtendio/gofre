package middleware

import (
	"bytes"
	"context"
	"github.com/ixtendio/gofre/router/path"

	"github.com/ixtendio/gofre/response"
	"net/http"
	"net/url"
	"testing"
)

func TestRequestDumper(t *testing.T) {
	r, _ := http.NewRequest("POST", "https://www.domain.com?qParam1=qVal1", bytes.NewBufferString(""))
	r.Form = url.Values{"param1": {"val1.1", "val1.2"}, "param2": {"val2.1", "val2.2"}}
	r.ContentLength = 256
	r.Header = http.Header{"h1": {"hv1.1"}, "h2": {"hv2.1", "hv2.2"}}
	r.RemoteAddr = "175.123.23.1"
	r.AddCookie(&http.Cookie{
		Name:  "Cookie1",
		Value: "CookieVal1",
	})
	var result string
	type args struct {
		req    *http.Request
		logger func(val string)
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "call",
			args: args{
				req: r,
				logger: func(val string) {
					result = val
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RequestDumper(tt.args.logger)(func(ctx context.Context, mc path.MatchingContext) (response.HttpResponse, error) {
				return response.PlainTextHttpResponseOK(""), nil
			})(context.Background(), path.MatchingContext{R: tt.args.req})
			if err != nil {
				t.Fatalf("RequestDumper() got err: %v", err)
			}
			if len(result) == 0 {
				t.Errorf("RequestDumper() got: %v, want not empty", result)
			}
		})
	}
}
