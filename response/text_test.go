package response

import (
	"github.com/ixtendio/gofre/router/path"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHttpTextResponse_Write(t *testing.T) {
	req := path.MatchingContext{R: &http.Request{}}
	type args struct {
		httpStatusCode int
		httpHeaders    HttpHeaders
		httpCookies    HttpCookies
		payload        string
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
				payload:        "hello",
			},
			want: want{
				httpCode:    201,
				httpHeaders: http.Header{"X-Content-Type-Options": {"nosniff"}},
				body:        []byte("hello"),
			},
			wantErr: false,
		},
		{
			name: "with custom headers",
			args: args{
				httpStatusCode: 202,
				httpHeaders:    HttpHeaders{"Content-Type": "test"},
			},
			want: want{
				httpCode:    202,
				httpHeaders: http.Header{"Content-Type": {"test"}, "X-Content-Type-Options": {"nosniff"}},
				body:        nil,
			},
			wantErr: false,
		},
		{
			name: "with custom cookies",
			args: args{
				httpStatusCode: 202,
				httpCookies: NewHttpCookie(&http.Cookie{
					Name:  "cookie1",
					Value: "val1",
				}, &http.Cookie{
					Name:  "cookie2",
					Value: "val2",
				}),
			},
			want: want{
				httpCode:    202,
				httpHeaders: http.Header{"X-Content-Type-Options": {"nosniff"}, "Set-Cookie": {"cookie1=val1", "cookie2=val2"}},
				body:        nil,
			},
			wantErr: false,
		},
		{
			name: "with all parameters",
			args: args{
				httpStatusCode: 202,
				payload:        "hello1",
				httpHeaders:    HttpHeaders{"header1": "val1"},
				httpCookies: NewHttpCookie(&http.Cookie{
					Name:  "cookie3",
					Value: "val3",
				}),
			},
			want: want{
				httpCode:    202,
				httpHeaders: http.Header{"X-Content-Type-Options": {"nosniff"}, "Set-Cookie": {"cookie3=val3"}, "Header1": {"val1"}},
				body:        []byte("hello1"),
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
			resp := &HttpTextResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: tt.args.httpStatusCode,
					HttpHeaders:    tt.args.httpHeaders,
					HttpCookies:    tt.args.httpCookies,
				},
				Payload: tt.args.payload,
			}
			responseRecorder := httptest.NewRecorder()
			err := resp.Write(responseRecorder, req)
			if tt.wantErr {
				if err == nil {
					t.Errorf("HttpTextResponse() want error but got nil")
				}
			} else {
				got := want{
					httpCode:    responseRecorder.Code,
					httpHeaders: responseRecorder.Header(),
					body:        responseRecorder.Body.Bytes(),
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("HttpTextResponse.Write() got:  %v, want: %v", got, tt.want)
				}
			}
		})
	}
}

func TestPlainTextHttpResponse(t *testing.T) {
	type args struct {
		statusCode int
		payload    string
	}
	tests := []struct {
		name string
		args args
		want *HttpTextResponse
	}{
		{
			name: "construct",
			args: args{
				statusCode: 201,
				payload:    "test",
			},
			want: &HttpTextResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 201,
					ContentType:    plainTextContentType,
				},
				Payload: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PlainTextHttpResponse(tt.args.statusCode, tt.args.payload); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PlainTextHttpResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlainTextHttpResponseOK(t *testing.T) {
	type args struct {
		payload string
	}
	tests := []struct {
		name string
		args args
		want *HttpTextResponse
	}{
		{
			name: "construct",
			args: args{
				payload: "test",
			},
			want: &HttpTextResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 200,
					ContentType:    plainTextContentType,
				},
				Payload: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PlainTextHttpResponseOK(tt.args.payload); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PlainTextHttpResponseOK() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlainTextHttpResponseWithHeaders(t *testing.T) {
	type args struct {
		statusCode int
		payload    string
		headers    HttpHeaders
	}
	tests := []struct {
		name string
		args args
		want *HttpTextResponse
	}{
		{
			name: "construct",
			args: args{
				statusCode: 201,
				headers:    HttpHeaders{"x-Header1": "val1", "x-Header2": "val2"},
				payload:    "test",
			},
			want: &HttpTextResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 201,
					ContentType:    plainTextContentType,
					HttpHeaders:    HttpHeaders{"x-Header1": "val1", "x-Header2": "val2"},
				},
				Payload: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PlainTextHttpResponseWithHeaders(tt.args.statusCode, tt.args.payload, tt.args.headers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PlainTextHttpResponseWithHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlainTextResponseWithHeadersAndCookies(t *testing.T) {
	cookies := NewHttpCookie(&http.Cookie{
		Name:  "cookie1",
		Value: "val1",
	}, &http.Cookie{
		Name:  "cookie2",
		Value: "val2",
	})
	type args struct {
		statusCode int
		payload    string
		headers    HttpHeaders
		cookies    HttpCookies
	}
	tests := []struct {
		name string
		args args
		want *HttpTextResponse
	}{
		{
			name: "construct",
			args: args{
				statusCode: 201,
				headers:    HttpHeaders{"x-Header1": "val1", "x-Header2": "val2"},
				cookies:    cookies,
				payload:    "test",
			},
			want: &HttpTextResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 201,
					ContentType:    plainTextContentType,
					HttpHeaders:    HttpHeaders{"x-Header1": "val1", "x-Header2": "val2"},
					HttpCookies:    cookies,
				},
				Payload: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PlainTextResponseWithHeadersAndCookies(tt.args.statusCode, tt.args.payload, tt.args.headers, tt.args.cookies); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PlainTextResponseWithHeadersAndCookies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHtmlHttpResponseOK(t *testing.T) {
	type args struct {
		payload string
	}
	tests := []struct {
		name string
		args args
		want *HttpTextResponse
	}{
		{
			name: "constructor",
			args: args{payload: "hello"},
			want: &HttpTextResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: http.StatusOK,
					ContentType:    htmlContentType,
				},
				Payload: "hello",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HtmlHttpResponseOK(tt.args.payload); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HtmlHttpResponseOK() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHtmlHttpResponse(t *testing.T) {
	type args struct {
		statusCode int
		payload    string
	}
	tests := []struct {
		name string
		args args
		want *HttpTextResponse
	}{
		{
			name: "constructor",
			args: args{
				statusCode: 400,
				payload:    "hello",
			},
			want: &HttpTextResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 400,
					ContentType:    htmlContentType,
				},
				Payload: "hello",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HtmlHttpResponse(tt.args.statusCode, tt.args.payload); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HtmlHttpResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHtmlHttpResponseWithHeaders(t *testing.T) {
	type args struct {
		statusCode int
		payload    string
		headers    HttpHeaders
	}
	tests := []struct {
		name string
		args args
		want *HttpTextResponse
	}{
		{
			name: "construct",
			args: args{
				statusCode: 201,
				headers:    HttpHeaders{"x-Header1": "val1", "x-Header2": "val2"},
				payload:    "test",
			},
			want: &HttpTextResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 201,
					ContentType:    htmlContentType,
					HttpHeaders:    HttpHeaders{"x-Header1": "val1", "x-Header2": "val2"},
				},
				Payload: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HtmlHttpResponseWithHeaders(tt.args.statusCode, tt.args.payload, tt.args.headers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HtmlHttpResponseWithHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHtmlResponseWithHeadersAndCookies(t *testing.T) {
	cookies := NewHttpCookie(&http.Cookie{
		Name:  "cookie1",
		Value: "val1",
	}, &http.Cookie{
		Name:  "cookie2",
		Value: "val2",
	})
	type args struct {
		statusCode int
		payload    string
		headers    HttpHeaders
		cookies    HttpCookies
	}
	tests := []struct {
		name string
		args args
		want *HttpTextResponse
	}{
		{
			name: "construct",
			args: args{
				statusCode: 201,
				headers:    HttpHeaders{"x-Header1": "val1", "x-Header2": "val2"},
				cookies:    cookies,
				payload:    "test",
			},
			want: &HttpTextResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 201,
					ContentType:    htmlContentType,
					HttpHeaders:    HttpHeaders{"x-Header1": "val1", "x-Header2": "val2"},
					HttpCookies:    cookies,
				},
				Payload: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HtmlResponseWithHeadersAndCookies(tt.args.statusCode, tt.args.payload, tt.args.headers, tt.args.cookies); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HtmlResponseWithHeadersAndCookies() = %v, want %v", got, tt.want)
			}
		})
	}
}
