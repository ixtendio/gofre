package response

import (
	"fmt"
	"github.com/ixtendio/gofre/router/path"

	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type mockTemplate struct {
}

func (t *mockTemplate) Execute(wr io.Writer, data any) error {
	_, err := wr.Write([]byte(fmt.Sprintf("template: data: %v", data)))
	return err
}

func (t *mockTemplate) ExecuteTemplate(wr io.Writer, name string, data any) error {
	_, err := wr.Write([]byte(fmt.Sprintf("template: %s, data: %v", name, data)))
	return err
}

var testTemplate = &mockTemplate{}
var htmlTemplate = &template.Template{}

func TestHttpTemplateResponse_Write(t *testing.T) {
	req := path.MatchingContext{R: &http.Request{}}
	type args struct {
		httpStatusCode int
		httpHeaders    HttpHeaders
		httpCookies    HttpCookies
		templateName   string
		templateData   any
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
			name: "valid output",
			args: args{
				httpStatusCode: 210,
				httpHeaders:    HttpHeaders{"Header1": "val1"},
				httpCookies: NewHttpCookie(&http.Cookie{
					Name:  "cookie3",
					Value: "val3",
				}),
				templateName: "not_found",
				templateData: map[string]string{"key1": "val1", "key2": "val2"},
			},
			want: want{
				httpCode:    210,
				httpHeaders: http.Header{"X-Content-Type-Options": {"nosniff"}, "Set-Cookie": {"cookie3=val3"}, "Header1": {"val1"}},
				body:        []byte(fmt.Sprintf("template: %s, data: %v", "not_found", map[string]string{"key1": "val1", "key2": "val2"})),
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
			resp := &HttpTemplateResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: tt.args.httpStatusCode,
					HttpHeaders:    tt.args.httpHeaders,
					HttpCookies:    tt.args.httpCookies,
				},
				Template: testTemplate,
				Name:     tt.args.templateName,
				Data:     tt.args.templateData,
			}
			responseRecorder := httptest.NewRecorder()
			err := resp.Write(responseRecorder, req)
			if tt.wantErr {
				if err == nil {
					t.Errorf("HttpTemplateResponse() want error but got nil")
				}
			} else {
				got := want{
					httpCode:    responseRecorder.Code,
					httpHeaders: responseRecorder.Header(),
					body:        responseRecorder.Body.Bytes(),
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("HttpTemplateResponse.Write() got:  %v, want: %v", got, tt.want)
				}
			}
		})
	}
}

func TestTemplateHttpResponseOK(t *testing.T) {
	type args struct {
		template     ExecutableTemplate
		templateName string
		templateData any
	}
	tests := []struct {
		name string
		args args
		want *HttpTemplateResponse
	}{
		{
			name: "constructor",
			args: args{
				template:     testTemplate,
				templateName: "index",
				templateData: "data",
			},
			want: &HttpTemplateResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 200,
					ContentType:    plainTextContentType,
				},
				Template: testTemplate,
				Name:     "index",
				Data:     "data",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TemplateHttpResponseOK(tt.args.template, tt.args.templateName, tt.args.templateData); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TemplateHttpResponseOK() got: %v, want: %v", got, tt.want)
			}
		})
	}
}

func TestTemplateHttpResponseNotFound(t *testing.T) {
	type args struct {
		template     ExecutableTemplate
		templateName string
		templateData any
	}
	tests := []struct {
		name string
		args args
		want *HttpTemplateResponse
	}{
		{
			name: "constructor",
			args: args{
				template:     testTemplate,
				templateName: "index",
				templateData: "data",
			},
			want: &HttpTemplateResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 404,
					ContentType:    plainTextContentType,
				},
				Template: testTemplate,
				Name:     "index",
				Data:     "data",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TemplateHttpResponseNotFound(tt.args.template, tt.args.templateName, tt.args.templateData); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TemplateHttpResponseNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateHttpResponseWithHeaders(t *testing.T) {
	type args struct {
		template     ExecutableTemplate
		statusCode   int
		templateName string
		templateData any
		headers      HttpHeaders
	}
	tests := []struct {
		name string
		args args
		want *HttpTemplateResponse
	}{
		{
			name: "constructor",
			args: args{
				template:     testTemplate,
				statusCode:   500,
				templateName: "index",
				templateData: "data",
				headers:      HttpHeaders{"x-Header1": "val1", "x-Header2": "val2"},
			},
			want: &HttpTemplateResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 500,
					ContentType:    plainTextContentType,
					HttpHeaders:    HttpHeaders{"x-Header1": "val1", "x-Header2": "val2"},
				},
				Template: testTemplate,
				Name:     "index",
				Data:     "data",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TemplateHttpResponseWithHeaders(tt.args.template, tt.args.statusCode, tt.args.templateName, tt.args.templateData, tt.args.headers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TemplateHttpResponseWithHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateHttpResponseWithCookies(t *testing.T) {
	cookies := NewHttpCookie(&http.Cookie{
		Name:  "cookie1",
		Value: "val1",
	}, &http.Cookie{
		Name:  "cookie2",
		Value: "val2",
	})
	type args struct {
		template     ExecutableTemplate
		statusCode   int
		templateName string
		templateData any
		cookies      HttpCookies
	}
	tests := []struct {
		name string
		args args
		want *HttpTemplateResponse
	}{
		{
			name: "constructor",
			args: args{
				template:     testTemplate,
				statusCode:   500,
				templateName: "index",
				templateData: "data",
				cookies:      cookies,
			},
			want: &HttpTemplateResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 500,
					ContentType:    plainTextContentType,
					HttpCookies:    cookies,
				},
				Template: testTemplate,
				Name:     "index",
				Data:     "data",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TemplateHttpResponseWithCookies(tt.args.template, tt.args.statusCode, tt.args.templateName, tt.args.templateData, tt.args.cookies); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TemplateHttpResponseWithCookies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateHttpResponseWithHeadersAndCookies(t *testing.T) {
	cookies := NewHttpCookie(&http.Cookie{
		Name:  "cookie1",
		Value: "val1",
	}, &http.Cookie{
		Name:  "cookie2",
		Value: "val2",
	})
	type args struct {
		template     ExecutableTemplate
		statusCode   int
		templateName string
		templateData any
		headers      HttpHeaders
		cookies      HttpCookies
	}
	tests := []struct {
		name string
		args args
		want *HttpTemplateResponse
	}{
		{
			name: "constructor",
			args: args{
				template:     testTemplate,
				statusCode:   500,
				templateName: "index",
				templateData: "data",
				headers:      HttpHeaders{"x-Header1": "val1", "x-Header2": "val2"},
				cookies:      cookies,
			},
			want: &HttpTemplateResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 500,
					ContentType:    plainTextContentType,
					HttpHeaders:    HttpHeaders{"x-Header1": "val1", "x-Header2": "val2"},
					HttpCookies:    cookies,
				},
				Template: testTemplate,
				Name:     "index",
				Data:     "data",
			},
		},
		{
			name: "with custom Content-Type",
			args: args{
				template:     testTemplate,
				statusCode:   500,
				templateName: "index",
				templateData: "data",
				headers:      HttpHeaders{"Content-Type": "something"},
			},
			want: &HttpTemplateResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 500,
					ContentType:    "something",
					HttpHeaders:    HttpHeaders{"Content-Type": "something"},
				},
				Template: testTemplate,
				Name:     "index",
				Data:     "data",
			},
		},
		{
			name: "with html template",
			args: args{
				template:     htmlTemplate,
				statusCode:   500,
				templateName: "index",
				templateData: "data",
			},
			want: &HttpTemplateResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 500,
					ContentType:    htmlContentType,
				},
				Template: htmlTemplate,
				Name:     "index",
				Data:     "data",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TemplateHttpResponseWithHeadersAndCookies(tt.args.template, tt.args.statusCode, tt.args.templateName, tt.args.templateData, tt.args.headers, tt.args.cookies); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TemplateHttpResponseWithHeadersAndCookies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNilExecutableTemplate(t1 *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "all methods should return nil"},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &NilTemplate{}
			if t.Execute(nil, nil) != nil {
				t1.Errorf("Execute() returned non nil, want nil")
				return
			}
			if t.ExecuteTemplate(nil, "", nil) != nil {
				t1.Errorf("ExecuteTemplate() returned non nil, want nil")
				return
			}
		})
	}
}
