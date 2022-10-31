package response

import (
	"fmt"
	"github.com/ixtendio/gofre/request"
	html "html/template"
	"io"
	"net/http"
)

const htmlContentType = "text/html; charset=utf-8"
const plainTextContentType = "text/plain; charset=utf-8"

type ExecutableTemplate interface {
	Execute(wr io.Writer, data any) error
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

// NilTemplate implements ExecutableTemplate and can be used when you use static resources without templating
type NilTemplate struct {
}

func (t NilTemplate) Execute(wr io.Writer, data any) error {
	return nil
}
func (t NilTemplate) ExecuteTemplate(wr io.Writer, name string, data any) error {
	return nil
}

type HttpTemplateResponse struct {
	HttpHeadersResponse
	Template ExecutableTemplate
	Name     string
	Data     any
}

func (r *HttpTemplateResponse) Write(w http.ResponseWriter, req *request.HttpRequest) error {
	// write the headers
	if err := r.HttpHeadersResponse.Write(w, req); err != nil {
		return err
	}

	if err := r.Template.ExecuteTemplate(w, r.Name, r.Data); err != nil {
		return fmt.Errorf("failed rendering the template: %s, err: %w", r.Name, err)
	}
	return nil
}

// TemplateHttpResponseOK creates a 200 success HTML response
func TemplateHttpResponseOK(template ExecutableTemplate, templateName string, templateData any) *HttpTemplateResponse {
	return TemplateHttpResponseWithHeadersAndCookies(template, http.StatusOK, templateName, templateData, nil, nil)
}

// TemplateHttpResponseNotFound creates a 404 HTML response
func TemplateHttpResponseNotFound(template ExecutableTemplate, templateName string, templateData any) *HttpTemplateResponse {
	return TemplateHttpResponseWithHeadersAndCookies(template, http.StatusNotFound, templateName, templateData, nil, nil)
}

// TemplateHttpResponseWithHeaders creates an HTML response with custom headers
func TemplateHttpResponseWithHeaders(template ExecutableTemplate, statusCode int, templateName string, templateData any, headers http.Header) *HttpTemplateResponse {
	return TemplateHttpResponseWithHeadersAndCookies(template, statusCode, templateName, templateData, headers, nil)
}

// TemplateHttpResponseWithCookies creates an HTML response with custom cookies
func TemplateHttpResponseWithCookies(template ExecutableTemplate, statusCode int, templateName string, templateData any, cookies []http.Cookie) *HttpTemplateResponse {
	return TemplateHttpResponseWithHeadersAndCookies(template, statusCode, templateName, templateData, nil, cookies)
}

// TemplateHttpResponseWithHeadersAndCookies creates an HTML response with custom headers and cookies
func TemplateHttpResponseWithHeadersAndCookies(template ExecutableTemplate, statusCode int, templateName string, templateData any, headers http.Header, cookies []http.Cookie) *HttpTemplateResponse {
	if headers == nil {
		headers = http.Header{}
	}
	if len(headers.Get("Content-Type")) == 0 {
		if _, ok := template.(*html.Template); ok {
			headers.Set("Content-Type", htmlContentType)
		} else {
			headers.Set("Content-Type", plainTextContentType)
		}
	}
	return &HttpTemplateResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    headers,
			HttpCookies:    NewHttpCookies(cookies),
		},
		Template: template,
		Name:     templateName,
		Data:     templateData,
	}
}
