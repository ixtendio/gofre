package router

import (
	"context"
	"fmt"
	"github.com/ixtendio/gofre/request"
	"github.com/ixtendio/gofre/response"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

type fakeResponseWriter struct {
	code    int
	headers http.Header
	payload []byte
}

func (m *fakeResponseWriter) Header() (h http.Header) {
	return m.headers
}

func (m *fakeResponseWriter) Write(p []byte) (n int, err error) {
	m.payload = p
	return len(p), nil
}

func (m *fakeResponseWriter) WriteHeader(code int) {
	m.code = code
}

func newFakeResponseWriter() *fakeResponseWriter {
	return &fakeResponseWriter{
		code:    0,
		headers: make(map[string][]string),
		payload: nil,
	}
}

func TestRouter_ServeHTTP(t *testing.T) {
	type args struct {
		writer *fakeResponseWriter
		req    *http.Request
	}
	type want struct {
		handlerInvoked  bool
		uriVars         map[string]string
		responseCode    int
		responseData    string
		responseHeaders http.Header
	}
	var wantRequest *request.HttpRequest
	okHandler := func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		wantRequest = r
		return response.PlainTextHttpResponseOK("ok"), nil
	}
	errorHandler := func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		wantRequest = r
		return nil, fmt.Errorf("a simple error")
	}
	tests := []struct {
		name   string
		args   args
		want   want
		router *Router
	}{
		{
			name: "root match",
			args: args{
				writer: newFakeResponseWriter(),
				req: &http.Request{
					Method: "GET",
					URL:    mustParseURL("/"),
				},
			},
			want: want{
				handlerInvoked: true,
				uriVars:        map[string]string{},
				responseCode:   200,
				responseData:   "ok",
				responseHeaders: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"},
					"X-Content-Type-Options": {"nosniff"}},
			},
			router: NewRouter(false, defaultErrLogFunc).
				Handle("GET", "/", okHandler),
		},
		{
			name: "match with vars",
			args: args{
				writer: newFakeResponseWriter(),
				req: &http.Request{
					Method: "GET",
					URL:    mustParseURL("/users/batman"),
				},
			},
			want: want{
				handlerInvoked: true,
				uriVars:        map[string]string{"userId": "batman"},
				responseCode:   200,
				responseData:   "ok",
				responseHeaders: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"},
					"X-Content-Type-Options": {"nosniff"}},
			},
			router: NewRouter(false, defaultErrLogFunc).
				Handle("GET", "/users/{userId}", okHandler).
				Handle("DELETE", "/users/{userId}", okHandler),
		},
		{
			name: "error handler",
			args: args{
				writer: newFakeResponseWriter(),
				req: &http.Request{
					Method: "POST",
					URL:    mustParseURL("/users/batman"),
				},
			},
			want: want{
				handlerInvoked:  true,
				uriVars:         map[string]string{"userId": "batman"},
				responseCode:    500,
				responseData:    "",
				responseHeaders: map[string][]string{},
			},
			router: NewRouter(false, defaultErrLogFunc).
				Handle("GET", "/users/{userId}", okHandler).
				Handle("POST", "/users/{userId}", errorHandler),
		},
		{
			name: "handler not match",
			args: args{
				writer: newFakeResponseWriter(),
				req: &http.Request{
					Method: "POST",
					URL:    mustParseURL("/users/batman/hello"),
				},
			},
			want: want{
				handlerInvoked:  false,
				uriVars:         map[string]string{},
				responseCode:    404,
				responseData:    "",
				responseHeaders: map[string][]string{},
			},
			router: NewRouter(false, defaultErrLogFunc).
				Handle("GET", "/users/{userId}", okHandler).
				Handle("POST", "/users/{userId}", errorHandler),
		},
	}
	for _, tt := range tests {
		wantRequest = nil
		t.Run(tt.name, func(t *testing.T) {
			tt.router.ServeHTTP(tt.args.writer, tt.args.req)
			if tt.want.handlerInvoked {
				if wantRequest == nil {
					t.Errorf("ServeHTTP() the request is null")
				}
				if !reflect.DeepEqual(wantRequest.UriVars, tt.want.uriVars) {
					t.Errorf("ServeHTTP() UriVars = %v, want %v", wantRequest.UriVars, tt.want.uriVars)
				}
			} else {
				if wantRequest != nil {
					t.Errorf("ServeHTTP() the request should be null")
				}
			}
			if tt.args.writer.code != tt.want.responseCode {
				t.Errorf("ServeHTTP() responseCode = %v, want %v", tt.args.writer.code, tt.want.responseCode)
			}
			if string(tt.args.writer.payload) != tt.want.responseData {
				t.Errorf("ServeHTTP() responseData = %v, want %v", string(tt.args.writer.payload), tt.want.responseData)
			}
			if !reflect.DeepEqual(tt.args.writer.headers, tt.want.responseHeaders) {
				t.Errorf("ServeHTTP() responseHeaders = %v, want %v", tt.args.writer.headers, tt.want.responseHeaders)
			}
		})
	}
}

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("Failed parsing the url: %s, err:%v", rawURL, err)
	}
	return u
}
