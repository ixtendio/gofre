package response

import (
	"github.com/ixtendio/gofre/router/path"
	"net/http"
)

// HttpRedirectResponse implements response.HttpResponse for writing a redirect URL
type HttpRedirectResponse struct {
	HttpHeadersResponse
	Url string
}

func (r *HttpRedirectResponse) Write(w http.ResponseWriter, mc path.MatchingContext) error {
	http.Redirect(w, mc.R, r.Url, r.HttpStatusCode)
	return nil
}

// RedirectHttpResponseMovedPermanently creates a redirect response with the status code 301
func RedirectHttpResponseMovedPermanently(url string) *HttpRedirectResponse {
	return redirectHttpResponse(http.StatusMovedPermanently, url)
}

// RedirectHttpResponse creates a redirect response with the status code 302
func RedirectHttpResponse(url string) *HttpRedirectResponse {
	return redirectHttpResponse(http.StatusFound, url)
}

// RedirectHttpResponseSeeOther creates a redirect response with the status code 303
func RedirectHttpResponseSeeOther(url string) *HttpRedirectResponse {
	return redirectHttpResponse(http.StatusSeeOther, url)
}

func redirectHttpResponse(statusCode int, url string) *HttpRedirectResponse {
	return &HttpRedirectResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: statusCode,
		},
		Url: url,
	}
}
