package response

import (
	"github.com/ixtendio/gofre/request"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHttpRedirectResponse_Write(t *testing.T) {
	req := request.HttpRequest{R: &http.Request{}}
	type args struct {
		url            string
		httpStatusCode int
	}
	type want struct {
		httpCode    int
		httpHeaders http.Header
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "write redirect",
			args: args{
				url:            "https://www.website.com",
				httpStatusCode: http.StatusMovedPermanently,
			},
			want: want{
				httpCode:    http.StatusMovedPermanently,
				httpHeaders: http.Header{"Location": {"https://www.website.com"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &HttpRedirectResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: tt.args.httpStatusCode,
				},
				Url: tt.args.url,
			}
			responseRecorder := httptest.NewRecorder()
			if err := resp.Write(responseRecorder, req); err != nil {
				t.Fatalf("HttpRedirectResponse.Write() got error: %v", err)
			}
			got := want{
				httpCode:    responseRecorder.Code,
				httpHeaders: responseRecorder.Header(),
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HttpRedirectResponse.Write() got:  %v, want: %v", got, tt.want)
			}
		})
	}
}

func TestRedirectHttpResponseMovedPermanently(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want *HttpRedirectResponse
	}{
		{
			name: "constructor",
			args: args{url: "https://www.website.com"},
			want: &HttpRedirectResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: http.StatusMovedPermanently,
				},
				Url: "https://www.website.com",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RedirectHttpResponseMovedPermanently(tt.args.url); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RedirectHttpResponseMovedPermanently() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedirectHttpResponse(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want *HttpRedirectResponse
	}{
		{
			name: "constructor",
			args: args{url: "https://www.website.com"},
			want: &HttpRedirectResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: http.StatusFound,
				},
				Url: "https://www.website.com",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RedirectHttpResponse(tt.args.url); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RedirectHttpResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedirectHttpResponseSeeOther(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want *HttpRedirectResponse
	}{
		{
			name: "constructor",
			args: args{url: "https://www.website.com"},
			want: &HttpRedirectResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: http.StatusSeeOther,
				},
				Url: "https://www.website.com",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RedirectHttpResponseSeeOther(tt.args.url); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RedirectHttpResponseSeeOther() = %v, want %v", got, tt.want)
			}
		})
	}
}
