package response

import (
	"errors"
	"github.com/ixtendio/gofre/router/path"

	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHttpJsonResponse_Write(t *testing.T) {
	req := path.MatchingContext{R: &http.Request{}}
	type args struct {
		httpStatusCode int
		httpHeaders    HttpHeaders
		httpCookies    HttpCookies
		payload        any
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
				body:        emptyJson,
			},
			wantErr: false,
		},
		{
			name: "with body",
			args: args{
				httpStatusCode: 201,
				payload:        map[string]string{"status": "ok"},
			},
			want: want{
				httpCode:    201,
				httpHeaders: http.Header{"X-Content-Type-Options": {"nosniff"}},
				body:        []byte(`{"status":"ok"}`),
			},
			wantErr: false,
		},
		{
			name: "with string body",
			args: args{
				httpStatusCode: 201,
				payload:        "hello",
			},
			want: want{
				httpCode:    201,
				httpHeaders: http.Header{"X-Content-Type-Options": {"nosniff"}},
				body:        []byte(`"hello"`),
			},
			wantErr: false,
		},
		{
			name: "with custom headers",
			args: args{
				httpStatusCode: 202,
				httpHeaders:    HttpHeaders{"Content-Type": jsonContentType},
			},
			want: want{
				httpCode:    202,
				httpHeaders: http.Header{"X-Content-Type-Options": {"nosniff"}, "Content-Type": {"application/json"}},
				body:        emptyJson,
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
				body:        emptyJson,
			},
			wantErr: false,
		},
		{
			name: "with all parameters",
			args: args{
				httpStatusCode: 202,
				payload:        map[string]string{"userId": "123"},
				httpHeaders:    HttpHeaders{"header1": "val1"},
				httpCookies: NewHttpCookie(&http.Cookie{
					Name:  "cookie3",
					Value: "val3",
				}),
			},
			want: want{
				httpCode:    202,
				httpHeaders: http.Header{"X-Content-Type-Options": {"nosniff"}, "Set-Cookie": {"cookie3=val3"}, "Header1": {"val1"}},
				body:        []byte(`{"userId":"123"}`),
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
			resp := &HttpJsonResponse{
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
					t.Errorf("HttpHeadersResponse() want error but got nil")
				}
			} else {
				got := want{
					httpCode:    responseRecorder.Code,
					httpHeaders: responseRecorder.Header(),
					body:        responseRecorder.Body.Bytes(),
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("HttpJsonResponse.Write() got:  %v, want: %v", got, tt.want)
				}
			}
		})
	}
}

func TestJsonHttpResponseOK(t *testing.T) {
	type args struct {
		payload any
	}
	tests := []struct {
		name string
		args args
		want *HttpJsonResponse
	}{
		{
			name: "constructor",
			args: args{payload: "hello"},
			want: &HttpJsonResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: http.StatusOK,
					ContentType:    jsonContentType,
				},
				Payload: "hello",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JsonHttpResponseOK(tt.args.payload); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsonHttpResponseOK() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJsonHttpResponse(t *testing.T) {
	type args struct {
		statusCode int
		payload    any
	}
	tests := []struct {
		name string
		args args
		want *HttpJsonResponse
	}{
		{
			name: "constructor",
			args: args{
				statusCode: 203,
				payload:    "hello",
			},
			want: &HttpJsonResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 203,
					ContentType:    jsonContentType,
				},
				Payload: "hello",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JsonHttpResponse(tt.args.statusCode, tt.args.payload); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsonHttpResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJsonHttpResponseWithCookies(t *testing.T) {
	cookies := NewHttpCookie(&http.Cookie{
		Name:  "cookie1",
		Value: "val1",
	}, &http.Cookie{
		Name:  "cookie2",
		Value: "val2",
	})
	type args struct {
		statusCode int
		payload    any
		cookies    HttpCookies
	}
	tests := []struct {
		name string
		args args
		want *HttpJsonResponse
	}{
		{
			name: "constructor",
			args: args{
				statusCode: 202,
				payload:    "hello",
				cookies:    cookies,
			},
			want: &HttpJsonResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 202,
					ContentType:    jsonContentType,
					HttpCookies:    cookies,
				},
				Payload: "hello",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JsonHttpResponseWithCookies(tt.args.statusCode, tt.args.payload, tt.args.cookies); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsonHttpResponseWithCookies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJsonHttpResponseWithHeaders(t *testing.T) {
	type args struct {
		statusCode int
		payload    any
		headers    HttpHeaders
	}
	tests := []struct {
		name string
		args args
		want *HttpJsonResponse
	}{
		{
			name: "constructor",
			args: args{
				statusCode: 202,
				payload:    "hello",
				headers:    HttpHeaders{"x-Header1": "val1", "x-Header2": "val2"},
			},
			want: &HttpJsonResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 202,
					ContentType:    jsonContentType,
					HttpHeaders:    HttpHeaders{"x-Header1": "val1", "x-Header2": "val2"},
				},
				Payload: "hello",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JsonHttpResponseWithHeaders(tt.args.statusCode, tt.args.payload, tt.args.headers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsonHttpResponseWithHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJsonHttpResponseWithHeadersAndCookies(t *testing.T) {
	cookies := NewHttpCookie(&http.Cookie{
		Name:  "cookie11",
		Value: "val1",
	}, &http.Cookie{
		Name:  "cookie22",
		Value: "val2",
	})
	type args struct {
		statusCode int
		payload    any
		headers    HttpHeaders
		cookies    HttpCookies
	}
	tests := []struct {
		name string
		args args
		want *HttpJsonResponse
	}{
		{
			name: "empty constructor",
			args: args{},
			want: &HttpJsonResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 0,
					ContentType:    jsonContentType,
				},
				Payload: nil,
			},
		},
		{
			name: "empty all args",
			args: args{
				statusCode: 205,
				payload:    "test1",
				headers:    HttpHeaders{"x-cust-val1": "val1", "x-cust-val2": "val2"},
				cookies:    cookies,
			},
			want: &HttpJsonResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 205,
					ContentType:    jsonContentType,
					HttpHeaders:    HttpHeaders{"x-cust-val1": "val1", "x-cust-val2": "val2"},
					HttpCookies:    cookies,
				},
				Payload: "test1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JsonHttpResponseWithHeadersAndCookies(tt.args.statusCode, tt.args.payload, tt.args.headers, tt.args.cookies); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsonHttpResponseWithHeadersAndCookies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJsonErrorHttpResponse(t *testing.T) {
	type args struct {
		statusCode int
		err        error
	}
	tests := []struct {
		name string
		args args
		want *HttpJsonResponse
	}{
		{
			name: "constructor",
			args: args{
				statusCode: 400,
				err:        errors.New("an error"),
			},
			want: &HttpJsonResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 400,
					ContentType:    jsonContentType,
				},
				Payload: map[string]string{"error": "An error"},
			},
		},
		{
			name: "constructor with empty error message",
			args: args{
				statusCode: 400,
				err:        errors.New(""),
			},
			want: &HttpJsonResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 400,
					ContentType:    jsonContentType,
				},
				Payload: map[string]string{"error": ""},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JsonErrorHttpResponse(tt.args.statusCode, tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsonErrorHttpResponse() got: %v, want: %v", got, tt.want)
			}
		})
	}
}

func TestJsonErrorHttpResponseWithCookies(t *testing.T) {
	cookies := NewHttpCookie(&http.Cookie{
		Name:  "cookie11",
		Value: "val1",
	}, &http.Cookie{
		Name:  "cookie22",
		Value: "val2",
	})
	type args struct {
		statusCode int
		err        error
		cookies    HttpCookies
	}
	tests := []struct {
		name string
		args args
		want *HttpJsonResponse
	}{
		{
			name: "constructor",
			args: args{
				statusCode: 400,
				err:        errors.New("an error"),
				cookies:    cookies,
			},
			want: &HttpJsonResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 400,
					ContentType:    jsonContentType,
					HttpCookies:    cookies,
				},
				Payload: map[string]string{"error": "An error"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JsonErrorHttpResponseWithCookies(tt.args.statusCode, tt.args.err, tt.args.cookies); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsonErrorHttpResponseWithCookies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJsonErrorHttpResponseWithHeaders(t *testing.T) {
	type args struct {
		statusCode int
		err        error
		headers    HttpHeaders
	}
	tests := []struct {
		name string
		args args
		want *HttpJsonResponse
	}{
		{
			name: "constructor",
			args: args{
				statusCode: 400,
				err:        errors.New("an error"),
				headers:    HttpHeaders{"x-cust-val3": "val3"},
			},
			want: &HttpJsonResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 400,
					ContentType:    jsonContentType,
					HttpHeaders:    HttpHeaders{"x-cust-val3": "val3"},
				},
				Payload: map[string]string{"error": "An error"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JsonErrorHttpResponseWithHeaders(tt.args.statusCode, tt.args.err, tt.args.headers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsonErrorHttpResponseWithHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJsonErrorHttpResponseWithHeadersAndCookies(t *testing.T) {
	cookies := NewHttpCookie(&http.Cookie{
		Name:  "cookie11",
		Value: "val1",
	}, &http.Cookie{
		Name:  "cookie22",
		Value: "val2",
	}, &http.Cookie{
		Name:  "cookie23",
		Value: "val3",
	})
	type args struct {
		statusCode int
		err        error
		headers    HttpHeaders
		cookies    HttpCookies
	}
	tests := []struct {
		name string
		args args
		want *HttpJsonResponse
	}{
		{
			name: "empty constructor",
			args: args{},
			want: &HttpJsonResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 0,
					ContentType:    jsonContentType,
				},
				Payload: map[string]string{"error": ""},
			},
		},
		{
			name: "constructor with all params",
			args: args{
				statusCode: 500,
				err:        errors.New("another error"),
				headers:    HttpHeaders{"x-cust-val12": "val12"},
				cookies:    cookies,
			},
			want: &HttpJsonResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 500,
					ContentType:    jsonContentType,
					HttpHeaders:    HttpHeaders{"x-cust-val12": "val12"},
					HttpCookies:    cookies,
				},
				Payload: map[string]string{"error": "Another error"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JsonErrorHttpResponseWithHeadersAndCookies(tt.args.statusCode, tt.args.err, tt.args.headers, tt.args.cookies); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsonErrorHttpResponseWithHeadersAndCookies() = %v, want %v", got, tt.want)
			}
		})
	}
}
