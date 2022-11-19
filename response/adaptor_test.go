package response

import (
	"github.com/ixtendio/gofre/router/path"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type testHandlerFunc struct {
	f func(http.ResponseWriter, *http.Request)
}

func (f testHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.f(w, r)
}

func TestHandlerAdaptor(t *testing.T) {
	req := path.MatchingContext{R: &http.Request{}}
	type args struct {
		handler     http.Handler
		httpHeaders HttpHeaders
	}
	type want struct {
		httpCode    int
		httpHeaders http.Header
		body        []byte
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "without custom headers",
			args: args{
				handler: testHandlerFunc{
					f: func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(201)
						w.Write([]byte("hello"))
					},
				},
				httpHeaders: nil,
			},
			want: want{
				httpCode:    201,
				httpHeaders: http.Header{},
				body:        []byte("hello"),
			},
		},
		{
			name: "with custom headers",
			args: args{
				handler: testHandlerFunc{
					f: func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(201)
						w.Write([]byte("hello"))
					},
				},
				httpHeaders: HttpHeaders{"Content-Type": "text/plain; charset=utf-8"},
			},
			want: want{
				httpCode:    201,
				httpHeaders: http.Header{"Content-Type": {"text/plain; charset=utf-8"}},
				body:        []byte("hello"),
			},
		},
	}
	for _, tt := range tests {
		responseRecorder := httptest.NewRecorder()
		t.Run(tt.name, func(t *testing.T) {
			adaptor := HandlerAdaptor(tt.args.handler)
			if tt.args.httpHeaders != nil {
				for k, v := range tt.args.httpHeaders {
					adaptor.Headers().Set(k, v)
				}
			}
			adaptor.Write(responseRecorder, req)
			got := want{
				httpCode:    responseRecorder.Code,
				httpHeaders: responseRecorder.Header(),
				body:        responseRecorder.Body.Bytes(),
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HandlerAdaptor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandlerFuncAdaptor(t *testing.T) {
	req := path.MatchingContext{R: &http.Request{}}
	type args struct {
		handler     http.HandlerFunc
		httpHeaders HttpHeaders
		httpCookies HttpCookies
	}
	type want struct {
		httpCode    int
		httpHeaders http.Header
		body        []byte
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "without custom headers",
			args: args{
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(201)
					w.Write([]byte("hello"))
				},
				httpHeaders: nil,
			},
			want: want{
				httpCode:    201,
				httpHeaders: http.Header{},
				body:        []byte("hello"),
			},
		},
		{
			name: "with custom headers",
			args: args{
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(201)
					w.Write([]byte("hello"))
				},
				httpHeaders: HttpHeaders{"Content-Type": "text/plain; charset=utf-8"},
			},
			want: want{
				httpCode:    201,
				httpHeaders: http.Header{"Content-Type": {"text/plain; charset=utf-8"}},
				body:        []byte("hello"),
			},
		},
		{
			name: "with custom headers and cookies",
			args: args{
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(201)
					w.Write([]byte("hello"))
				},
				httpHeaders: HttpHeaders{"Content-Type": "text/plain; charset=utf-8"},
				httpCookies: NewHttpCookie(&http.Cookie{
					Name:  "cookie1",
					Value: "val1",
				}),
			},
			want: want{
				httpCode:    201,
				httpHeaders: http.Header{"Content-Type": {"text/plain; charset=utf-8"}, "Set-Cookie": {"cookie1=val1"}},
				body:        []byte("hello"),
			},
		},
	}
	for _, tt := range tests {
		responseRecorder := httptest.NewRecorder()
		t.Run(tt.name, func(t *testing.T) {
			adaptor := HandlerFuncAdaptor(tt.args.handler)
			if tt.args.httpHeaders != nil {
				for k, v := range tt.args.httpHeaders {
					adaptor.Headers().Set(k, v)
				}
			}

			for _, k := range tt.args.httpCookies {
				adaptor.Cookies().Add(k)
			}

			adaptor.Write(responseRecorder, req)
			got := want{
				httpCode:    responseRecorder.Code,
				httpHeaders: responseRecorder.Header(),
				body:        responseRecorder.Body.Bytes(),
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HandlerAdaptor() = %v, want %v", got, tt.want)
			}
			if adaptor.StatusCode() != 0 {
				t.Errorf("HandlerAdaptor().StatusCode() should return 0 but got: %v", adaptor.StatusCode())
			}
		})
	}
}
