package response

import (
	"fmt"
	"github.com/ixtendio/gofre/router/path"

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

// HttpTemplateResponse implements response.HttpResponse and renders a template as a response
type HttpTemplateResponse struct {
	HttpHeadersResponse
	Template ExecutableTemplate
	Name     string
	Data     any
}

func (r *HttpTemplateResponse) Write(w http.ResponseWriter, mc path.MatchingContext) error {
	// write the headers
	if err := r.HttpHeadersResponse.Write(w, mc); err != nil {
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
// The headers, if present, once will be written to output will be added in the pool for re-use
func TemplateHttpResponseWithHeaders(template ExecutableTemplate, statusCode int, templateName string, templateData any, headers HttpHeaders) *HttpTemplateResponse {
	return TemplateHttpResponseWithHeadersAndCookies(template, statusCode, templateName, templateData, headers, nil)
}

// TemplateHttpResponseWithCookies creates an HTML response with custom cookies
// The cookies, if present, once will be written to output will be added in the pool for re-use
func TemplateHttpResponseWithCookies(template ExecutableTemplate, statusCode int, templateName string, templateData any, cookies HttpCookies) *HttpTemplateResponse {
	return TemplateHttpResponseWithHeadersAndCookies(template, statusCode, templateName, templateData, nil, cookies)
}

// TemplateHttpResponseWithHeadersAndCookies creates an HTML response with custom headers and cookies
// The headers and cookies, if present, once will be written to output will be added in the pool for re-use
func TemplateHttpResponseWithHeadersAndCookies(template ExecutableTemplate, statusCode int, templateName string, templateData any, headers HttpHeaders, cookies HttpCookies) *HttpTemplateResponse {
	var contentType string
	if headers != nil && len(headers[HeaderContentType]) > 0 {
		contentType = headers[HeaderContentType]
	} else {
		if _, ok := template.(*html.Template); ok {
			contentType = htmlContentType
		} else {
			contentType = plainTextContentType
		}
	}
	return &HttpTemplateResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			ContentType:    contentType,
			HttpHeaders:    headers,
			HttpCookies:    cookies,
		},
		Template: template,
		Name:     templateName,
		Data:     templateData,
	}
}
