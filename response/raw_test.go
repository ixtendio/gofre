package response

import (
	"github.com/ixtendio/gofre/router/path"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var testRawWriterFunc = func(w io.Writer) error {
	return nil
}

func TestRawWriterHttpResponse(t *testing.T) {
	type args struct {
		contentType string
		writeFunc   RawWriterFunc
	}
	tests := []struct {
		name string
		args args
		want *HttpRawResponse
	}{
		{
			name: "constructor",
			args: args{
				writeFunc:   testRawWriterFunc,
				contentType: "bytes",
			},
			want: &HttpRawResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 200,
					ContentType:    "bytes",
				},
				WriteFunc: testRawWriterFunc,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RawWriterHttpResponse(tt.args.contentType, tt.args.writeFunc)
			if got.WriteFunc == nil {
				t.Errorf("RawWriterHttpResponse() WriteFunc is nil, want not nil")
			}
			if !reflect.DeepEqual(got.HttpHeadersResponse, tt.want.HttpHeadersResponse) {
				t.Errorf("RawWriterHttpResponse().HttpHeadersResponse = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawWriterHttpResponseWithHeaders(t *testing.T) {
	type args struct {
		statusCode  int
		contentType string
		headers     HttpHeaders
		writeFunc   RawWriterFunc
	}
	tests := []struct {
		name string
		args args
		want *HttpRawResponse
	}{
		{
			name: "constructor",
			args: args{
				statusCode:  201,
				writeFunc:   testRawWriterFunc,
				contentType: "bytes",
				headers:     HttpHeaders{"h1": "v1"},
			},
			want: &HttpRawResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 201,
					ContentType:    "bytes",
					HttpHeaders:    HttpHeaders{"h1": "v1"},
				},
				WriteFunc: testRawWriterFunc,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RawWriterHttpResponseWithHeaders(tt.args.statusCode, tt.args.contentType, tt.args.headers, tt.args.writeFunc)
			if got.WriteFunc == nil {
				t.Errorf("RawWriterHttpResponseWithHeaders() WriteFunc is nil, want not nil")
			}
			if !reflect.DeepEqual(got.HttpHeadersResponse, tt.want.HttpHeadersResponse) {
				t.Errorf("RawWriterHttpResponseWithHeaders().HttpHeadersResponse = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawWriterHttpResponseWithCookies(t *testing.T) {
	cookies := NewHttpCookie(&http.Cookie{
		Name:  "cookie3",
		Value: "val3",
	})
	type args struct {
		statusCode  int
		contentType string
		cookies     HttpCookies
		writeFunc   RawWriterFunc
	}
	tests := []struct {
		name string
		args args
		want *HttpRawResponse
	}{
		{
			name: "constructor",
			args: args{
				statusCode:  201,
				writeFunc:   testRawWriterFunc,
				contentType: "bytes",
				cookies:     cookies,
			},
			want: &HttpRawResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 201,
					ContentType:    "bytes",
					HttpCookies:    cookies,
				},
				WriteFunc: testRawWriterFunc,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RawWriterHttpResponseWithCookies(tt.args.statusCode, tt.args.contentType, tt.args.cookies, tt.args.writeFunc)
			if got.WriteFunc == nil {
				t.Errorf("RawWriterHttpResponseWithCookies() WriteFunc is nil, want not nil")
			}
			if !reflect.DeepEqual(got.HttpHeadersResponse, tt.want.HttpHeadersResponse) {
				t.Errorf("RawWriterHttpResponseWithCookies().HttpHeadersResponse = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawWriterHttpResponseWithHeadersAndCookies(t *testing.T) {
	cookies := NewHttpCookie(&http.Cookie{
		Name:  "cookie4",
		Value: "val4",
	})
	type args struct {
		statusCode  int
		contentType string
		headers     HttpHeaders
		cookies     HttpCookies
		writeFunc   RawWriterFunc
	}
	tests := []struct {
		name string
		args args
		want *HttpRawResponse
	}{
		{
			name: "constructor",
			args: args{
				statusCode:  201,
				writeFunc:   testRawWriterFunc,
				contentType: "bytes",
				headers:     HttpHeaders{"h2": "v2"},
				cookies:     cookies,
			},
			want: &HttpRawResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 201,
					ContentType:    "bytes",
					HttpHeaders:    HttpHeaders{"h2": "v2"},
					HttpCookies:    cookies,
				},
				WriteFunc: testRawWriterFunc,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RawWriterHttpResponseWithHeadersAndCookies(tt.args.statusCode, tt.args.contentType, tt.args.headers, tt.args.cookies, tt.args.writeFunc)
			if got.WriteFunc == nil {
				t.Errorf("RawWriterHttpResponseWithHeadersAndCookies() WriteFunc is nil, want not nil")
			}
			if !reflect.DeepEqual(got.HttpHeadersResponse, tt.want.HttpHeadersResponse) {
				t.Errorf("RawWriterHttpResponseWithHeadersAndCookies().HttpHeadersResponse = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpRawResponse_Write(t *testing.T) {
	req := path.MatchingContext{R: &http.Request{}}
	type args struct {
		httpStatusCode int
		httpHeaders    HttpHeaders
		httpCookies    HttpCookies
		writeFunc      RawWriterFunc
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
				writeFunc: func(w io.Writer) error {
					_, err := w.Write([]byte("hello"))
					if err != nil {
						return err
					}
					return nil
				},
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
				httpCookies: NewHttpCookie(&http.Cookie{
					Name:  "cookie1",
					Value: "val1",
				}, &http.Cookie{
					Name:  "cookie2",
					Value: "val2",
				}),
				writeFunc: func(w io.Writer) error {
					_, err := w.Write([]byte("hello"))
					if err != nil {
						return err
					}
					return nil
				},
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
			resp := &HttpRawResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: tt.args.httpStatusCode,
					HttpHeaders:    tt.args.httpHeaders,
					HttpCookies:    tt.args.httpCookies,
				},
				WriteFunc: tt.args.writeFunc,
			}
			responseRecorder := httptest.NewRecorder()
			err := resp.Write(responseRecorder, req)
			if tt.wantErr {
				if err == nil {
					t.Errorf("HttpRawResponse() want error but got nil")
				}
			} else {
				got := want{
					httpCode:    responseRecorder.Code,
					httpHeaders: responseRecorder.Header(),
					body:        responseRecorder.Body.Bytes(),
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("HttpRawResponse.Write() got:  %v, want: %v", got, tt.want)
				}
			}
		})
	}
}
