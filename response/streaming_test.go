package response

import (
	"bytes"
	"github.com/ixtendio/gofre/router/path"

	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var testStreamReader = bytes.NewBufferString("")

func TestStreamHttpResponse(t *testing.T) {
	type args struct {
		reader      io.Reader
		contentType string
	}
	tests := []struct {
		name string
		args args
		want *HttpStreamResponse
	}{
		{
			name: "constructor",
			args: args{
				reader:      testStreamReader,
				contentType: "bytes",
			},
			want: &HttpStreamResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 200,
					ContentType:    "bytes",
				},
				Reader: testStreamReader,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StreamHttpResponse(tt.args.contentType, tt.args.reader); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StreamHttpResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStreamHttpResponseWithHeaders(t *testing.T) {
	type args struct {
		statusCode  int
		reader      io.Reader
		contentType string
		headers     HttpHeaders
	}
	tests := []struct {
		name string
		args args
		want *HttpStreamResponse
	}{
		{
			name: "constructor",
			args: args{
				statusCode:  201,
				reader:      testStreamReader,
				contentType: "bytes",
				headers:     HttpHeaders{"h1": "v1"},
			},
			want: &HttpStreamResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 201,
					ContentType:    "bytes",
					HttpHeaders:    HttpHeaders{"h1": "v1"},
					HttpCookies:    nil,
				},
				Reader: testStreamReader,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StreamHttpResponseWithHeaders(tt.args.statusCode, tt.args.contentType, tt.args.headers, tt.args.reader); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StreamHttpResponseWithHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStreamHttpResponseWithCookies(t *testing.T) {
	cookies := NewHttpCookie(&http.Cookie{
		Name:  "cookie3",
		Value: "val3",
	})
	type args struct {
		statusCode  int
		reader      io.Reader
		contentType string
		cookies     HttpCookies
	}
	tests := []struct {
		name string
		args args
		want *HttpStreamResponse
	}{
		{
			name: "constructor",
			args: args{
				statusCode:  201,
				reader:      testStreamReader,
				contentType: "bytes",
				cookies:     cookies,
			},
			want: &HttpStreamResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 201,
					ContentType:    "bytes",
					HttpCookies:    cookies,
				},
				Reader: testStreamReader,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StreamHttpResponseWithCookies(tt.args.statusCode, tt.args.contentType, tt.args.cookies, tt.args.reader); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StreamHttpResponseWithCookies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStreamHttpResponseWithHeadersAndCookies(t *testing.T) {
	cookies := NewHttpCookie(&http.Cookie{
		Name:  "cookie4",
		Value: "val4",
	})
	type args struct {
		statusCode  int
		reader      io.Reader
		contentType string
		headers     HttpHeaders
		cookies     HttpCookies
	}
	tests := []struct {
		name string
		args args
		want *HttpStreamResponse
	}{
		{
			name: "constructor",
			args: args{
				statusCode:  201,
				reader:      testStreamReader,
				contentType: "bytes",
				headers:     HttpHeaders{"h2": "v2"},
				cookies:     cookies,
			},
			want: &HttpStreamResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 201,
					ContentType:    "bytes",
					HttpHeaders:    HttpHeaders{"h2": "v2"},
					HttpCookies:    cookies,
				},
				Reader: testStreamReader,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StreamHttpResponseWithHeadersAndCookies(tt.args.statusCode, tt.args.contentType, tt.args.headers, tt.args.cookies, tt.args.reader); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StreamHttpResponseWithHeadersAndCookies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpStreamResponse_Write(t *testing.T) {
	cookies := NewHttpCookie(&http.Cookie{
		Name:  "cookie1",
		Value: "val1",
	}, &http.Cookie{
		Name:  "cookie2",
		Value: "val2",
	})
	req := path.MatchingContext{R: &http.Request{}}
	type args struct {
		httpStatusCode int
		httpHeaders    HttpHeaders
		httpCookies    HttpCookies
		payload        io.Reader
	}
	type want struct {
		httpCode    int
		httpHeaders http.Header
		body        []byte
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "without body",
			args: args{
				httpStatusCode: 201,
			},
			want: want{
				httpCode:    201,
				httpHeaders: http.Header{"X-Content-Type-Options": {"nosniff"}},
				body:        nil,
			},
			wantErr: false,
		},
		{
			name: "with body",
			args: args{
				httpStatusCode: 201,
				payload:        bytes.NewBufferString("hello"),
			},
			want: want{
				httpCode:    201,
				httpHeaders: http.Header{"X-Content-Type-Options": {"nosniff"}},
				body:        []byte("hello"),
			},
			wantErr: false,
		},
		{
			name: "with body headers and cookies",
			args: args{
				httpStatusCode: 201,
				httpHeaders:    HttpHeaders{"Content-Type": "bytes"},
				httpCookies:    cookies,
				payload:        bytes.NewBufferString("hello"),
			},
			want: want{
				httpCode:    201,
				httpHeaders: http.Header{"X-Content-Type-Options": {"nosniff"}, "Content-Type": {"bytes"}, "Set-Cookie": {"cookie1=val1", "cookie2=val2"}},
				body:        []byte("hello"),
			},
			wantErr: false,
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
			resp := &HttpStreamResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: tt.args.httpStatusCode,
					HttpCookies:    tt.args.httpCookies,
					HttpHeaders:    tt.args.httpHeaders,
				},
				Reader: tt.args.payload,
			}
			responseRecorder := httptest.NewRecorder()
			err := resp.Write(responseRecorder, req)
			if tt.wantErr {
				if err == nil {
					t.Errorf("HttpStreamResponse() want error but got nil")
				}
			} else {
				got := want{
					httpCode:    responseRecorder.Code,
					httpHeaders: responseRecorder.Header(),
					body:        responseRecorder.Body.Bytes(),
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("HttpStreamResponse.Write() got:  %v, want: %v", got, tt.want)
				}
			}
		})
	}
}
