package response

import (
	"fmt"
	"github.com/ixtendio/gow/request"
	"html/template"
	"net/http"
)

var defaultHtmlHeaders = map[string]string{
	"Content-Type": "text/html; charset=utf-8",
}

type HttpTemplateResponse struct {
	HttpHeadersResponse
	Template *template.Template
	Name     string
	Data     any
}

func (r *HttpTemplateResponse) Write(w http.ResponseWriter, reqContext *request.HttpRequest) error {
	// write the headers
	if err := r.HttpHeadersResponse.Write(w, reqContext); err != nil {
		return err
	}

	if err := r.Template.ExecuteTemplate(w, r.Name, r.Data); err != nil {
		return fmt.Errorf("failed rendering the template: %s, err: %w", r.Name, err)
	}
	return nil
}

// TemplateHttpResponseOK creates a 200 success HTML response
func TemplateHttpResponseOK(template *template.Template, templateName string, templateData any) *HttpTemplateResponse {
	return &HttpTemplateResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: http.StatusOK,
			HttpHeaders:    defaultHtmlHeaders,
			HttpCookies:    nil,
		},
		Template: template,
		Name:     templateName,
		Data:     templateData,
	}
}

// TemplateHttpResponseNotFound creates a 404 HTML response
func TemplateHttpResponseNotFound(template *template.Template, templateName string, templateData any) *HttpTemplateResponse {
	return &HttpTemplateResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: http.StatusNotFound,
			HttpHeaders:    defaultHtmlHeaders,
			HttpCookies:    nil,
		},
		Template: template,
		Name:     templateName,
		Data:     templateData,
	}
}

// TemplateHttpResponseWithHeaders creates an HTML response with custom headers
func TemplateHttpResponseWithHeaders(template *template.Template, statusCode int, templateName string, templateData any, headers map[string]string) *HttpTemplateResponse {
	headers["Content-Type"] = "text/html; charset=utf-8"
	return &HttpTemplateResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    headers,
			HttpCookies:    nil,
		},
		Template: template,
		Name:     templateName,
		Data:     templateData,
	}
}

// TemplateHttpResponseWithCookies creates an HTML response with custom cookies
func TemplateHttpResponseWithCookies(template *template.Template, statusCode int, templateName string, templateData any, cookies []*http.Cookie) *HttpTemplateResponse {
	return &HttpTemplateResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    defaultHtmlHeaders,
			HttpCookies:    cookies,
		},
		Template: template,
		Name:     templateName,
		Data:     templateData,
	}
}

// TemplateHttpResponseWithHeadersAndCookies creates an HTML response with custom headers and cookies
func TemplateHttpResponseWithHeadersAndCookies(template *template.Template, statusCode int, templateName string, templateData any, headers map[string]string, cookies []*http.Cookie) *HttpTemplateResponse {
	headers["Content-Type"] = "text/html; charset=utf-8"
	return &HttpTemplateResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
			HttpHeaders:    headers,
			HttpCookies:    cookies,
		},
		Template: template,
		Name:     templateName,
		Data:     templateData,
	}
}
