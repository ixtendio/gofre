package response

import (
	"github.com/ixtendio/gofre/router/path"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHttpHeadersResponse_Write(t *testing.T) {
	req := path.MatchingContext{R: &http.Request{}}
	type args struct {
		httpStatusCode int
		httpHeaders    HttpHeaders
		httpCookies    HttpCookies
	}
	type want struct {
		httpCode    int
		httpHeaders http.Header
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "without cookies",
			args: args{
				httpStatusCode: 200,
				httpHeaders:    HttpHeaders{"Content-Type": "text/plain; charset=utf-8"},
			},
			want: want{
				httpCode:    200,
				httpHeaders: http.Header{"X-Content-Type-Options": {"nosniff"}, "Content-Type": {"text/plain; charset=utf-8"}},
			},
		},
		{
			name: "with cookies",
			args: args{
				httpStatusCode: 200,
				httpHeaders:    HttpHeaders{"Content-Type": "text/plain; charset=utf-8"},
				httpCookies: NewHttpCookie(&http.Cookie{
					Name:  "cookie1",
					Value: "val1",
				}),
			},
			want: want{
				httpCode:    200,
				httpHeaders: http.Header{"X-Content-Type-Options": {"nosniff"}, "Content-Type": {"text/plain; charset=utf-8"}, "Set-Cookie": {"cookie1=val1"}},
			},
		},
		{
			name: "with status code 0",
			args: args{
				httpStatusCode: 0,
			},
			want: want{
				httpCode:    200,
				httpHeaders: http.Header{"X-Content-Type-Options": {"nosniff"}},
			},
		},
		{
			name: "with status code 1",
			args: args{
				httpStatusCode: 1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &HttpHeadersResponse{
				HttpStatusCode: tt.args.httpStatusCode,
				HttpHeaders:    tt.args.httpHeaders,
				HttpCookies:    tt.args.httpCookies,
			}
			responseRecorder := httptest.NewRecorder()
			err := resp.Write(responseRecorder, req)
			if tt.wantErr {
				if err == nil {
					t.Errorf("HttpHeadersResponse() want error but got nil")
				}
			} else {
				got := want{
					httpCode:    responseRecorder.Code,
					httpHeaders: responseRecorder.Header(),
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("HttpHeadersResponse() got:  %v, want: %v", got, tt.want)
				}
			}
		})
	}
}

func TestInternalServerErrorHttpResponse(t *testing.T) {
	tests := []struct {
		name string
		want *HttpHeadersResponse
	}{
		{
			name: "constructor",
			want: &HttpHeadersResponse{HttpStatusCode: http.StatusInternalServerError},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InternalServerErrorHttpResponse(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InternalServerErrorHttpResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}
