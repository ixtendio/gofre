package middleware

import (
	"context"
	"github.com/ixtendio/gofre/errors"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/request"
	"github.com/ixtendio/gofre/response"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	contentTypeValueTextPlain                   = "text/plain"
	contentTypeValueFormData                    = "multipart/form-data"
	contentTypeValueFormUrlencodedData          = "application/x-www-form-urlencoded"
	requestHeaderContentType                    = "Content-Type"
	requestHeaderOrigin                         = "Origin"
	requestHeaderAccessControlRequestMethod     = "Access-Control-Request-Method"
	requestHeaderAccessControlRequestHeaders    = "Access-Control-Request-Headers"
	responseHeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	responseHeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	responseHeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	responseHeaderAccessControlMaxAge           = "Access-Control-Max-Age"
	responseHeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	responseHeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	varyHeader                                  = "vary"
)

type CorsConfig struct {
	// determines if any origin is allowed to make request
	AnyOriginAllowed bool
	// this flag should be true when the resource supports user credentials in the request and false otherwise.
	SupportsCredentials bool
	// indicates how long the results of a pre-flight request can be cached in a pre-flight result cache
	PreflightMaxAgeSeconds int
	// a list of exposed headers that the resource might use and can be exposed (can be empty)
	ExposedHeaders []string
	// a list of methods that are supported by the resource (can be empty)
	AllowedHttpMethods []string
	// a list of headers that are supported by the resource (can be empty)
	AllowedHttpHeaders []string
	// a list of origins that are allowed access to the resource (can be empty)
	AllowedOrigins []string
}

func (c *CorsConfig) containsAllowedMethod(method string) bool {
	for _, m := range c.AllowedHttpMethods {
		if m == method {
			return true
		}
	}
	return false
}

func (c *CorsConfig) containsAllowedHeaderCaseInsensitive(header string) bool {
	for _, h := range c.AllowedHttpHeaders {
		if strings.EqualFold(h, header) {
			return true
		}
	}
	return false
}

func (c *CorsConfig) containsAllowedOrigin(origin string) bool {
	for _, o := range c.AllowedOrigins {
		if o == origin {
			return true
		}
	}
	return false
}

type corsRequestType int

const (
	simpleCorsRequestType corsRequestType = iota
	actualCorsRequestType
	preFlightCorsRequestType
	notCorsRequestType
	invalidCorsRequestType
)

// Cors enable client-side cross-origin requests by implementing W3C's CORS (Cross-Origin Resource Sharing) specification for resources
// This function is a transcription of Java code org.apache.catalina.filters.CorsFilter
func Cors(config CorsConfig) Middleware {
	return func(handler handler.Handler) handler.Handler {
		return func(ctx context.Context, req *request.HttpRequest) (response.HttpResponse, error) {
			httpResponse, err := handler(ctx, req)
			if err != nil {
				return nil, err
			}
			switch getRequestType(req.R) {
			case simpleCorsRequestType, actualCorsRequestType:
				if err := addSimpleCorsHeaders(req.R, httpResponse.Headers(), config); err != nil {
					return nil, err
				}
				return httpResponse, nil
			case preFlightCorsRequestType:
				if err := addPreFlightCorsHeaders(req.R, httpResponse.Headers(), config); err != nil {
					return nil, err
				}
				return httpResponse, nil
			case notCorsRequestType:
				addStandardCorsHeaders(req.R, httpResponse.Headers(), config)
				return httpResponse, nil
			default:
				return nil, errors.ErrDenied
			}
		}
	}
}

func addSimpleCorsHeaders(r *http.Request, responseHeaders http.Header, config CorsConfig) error {
	method := r.Method
	origin := r.Header.Get(requestHeaderOrigin)

	// Section 6.1.2
	if !isOriginAllowed(origin, config) {
		return errors.ErrDenied
	}

	if !config.containsAllowedMethod(method) {
		return errors.ErrDenied
	}

	addStandardCorsHeaders(r, responseHeaders, config)
	return nil
}

func addPreFlightCorsHeaders(r *http.Request, responseHeaders http.Header, config CorsConfig) error {
	origin := r.Header.Get(requestHeaderOrigin)

	// Section 6.2.2
	if !isOriginAllowed(origin, config) {
		return errors.ErrDenied
	}

	// Section 6.2.3
	if _, found := r.Header[requestHeaderAccessControlRequestMethod]; !found {
		return errors.ErrDenied
	}

	// Section 6.2.5
	accessControlRequestMethod := strings.TrimSpace(r.Header.Get(requestHeaderAccessControlRequestMethod))
	if !config.containsAllowedMethod(accessControlRequestMethod) {
		return errors.ErrDenied
	}

	// Section 6.2.4
	accessControlRequestHeadersHeader := strings.TrimSpace(r.Header.Get(requestHeaderAccessControlRequestHeaders))
	for _, h := range strings.Split(accessControlRequestHeadersHeader, ",") {
		h = strings.TrimSpace(h)
		if !config.containsAllowedHeaderCaseInsensitive(strings.TrimSpace(h)) {
			return errors.ErrDenied
		}
	}

	addStandardCorsHeaders(r, responseHeaders, config)
	return nil
}

func addStandardCorsHeaders(r *http.Request, responseHeaders http.Header, config CorsConfig) {
	method := r.Method
	origin := r.Header.Get(requestHeaderOrigin)

	// Local copy to avoid concurrency issues if isAnyOriginAllowed()
	// is overridden.
	if config.AnyOriginAllowed {
		responseHeaders.Set(responseHeaderAccessControlAllowOrigin, "*")
	} else {
		responseHeaders.Set(responseHeaderAccessControlAllowOrigin, origin)
		addVaryHeader(responseHeaders, requestHeaderOrigin)
	}

	if config.SupportsCredentials {
		responseHeaders.Set(responseHeaderAccessControlAllowCredentials, "true")
	}

	if len(config.ExposedHeaders) > 0 {
		responseHeaders.Set(responseHeaderAccessControlExposeHeaders, strings.Join(config.ExposedHeaders, ","))
	}

	if method == http.MethodOptions {
		addVaryHeader(responseHeaders, requestHeaderAccessControlRequestMethod)
		addVaryHeader(responseHeaders, requestHeaderAccessControlRequestHeaders)

		if config.PreflightMaxAgeSeconds > 0 {
			responseHeaders.Set(responseHeaderAccessControlMaxAge, strconv.Itoa(config.PreflightMaxAgeSeconds))
		}

		if len(config.AllowedHttpMethods) > 0 {
			responseHeaders.Set(responseHeaderAccessControlAllowMethods, strings.Join(config.AllowedHttpMethods, ","))
		}

		if len(config.AllowedHttpHeaders) > 0 {
			responseHeaders.Set(responseHeaderAccessControlAllowHeaders, strings.Join(config.AllowedHttpHeaders, ","))
		}
	}
}

func addVaryHeader(responseHeaders http.Header, name string) {
	var varyHeaders []string
	for _, vh := range responseHeaders.Values(varyHeader) {
		if vh != "" {
			for _, v := range strings.Split(vh, ",") {
				varyHeaders = append(varyHeaders, strings.TrimSpace(v))
			}
		}
	}
	if name == "*" || len(varyHeaders) == 0 {
		responseHeaders.Set(varyHeader, name)
		return
	}

	if len(varyHeaders) == 1 && strings.TrimSpace(varyHeaders[0]) == "*" {
		// No need to add any additional field
		return
	}

	for _, vh := range varyHeaders {
		if strings.TrimSpace(vh) == "*" {
			// '*' has been added without removing other values. Optimise.
			responseHeaders.Set(varyHeader, "*")
			return
		}
	}

	varyHeaders = append(varyHeaders, name)
	responseHeaders.Set(varyHeader, strings.Join(varyHeaders, ","))
}

func isOriginAllowed(originHeader string, config CorsConfig) bool {
	if config.AnyOriginAllowed {
		return true
	}

	// 'Origin' header is a case-sensitive match
	return config.containsAllowedOrigin(originHeader)
}

func getRequestType(r *http.Request) corsRequestType {
	if _, found := r.Header[requestHeaderOrigin]; !found {
		return notCorsRequestType
	}

	originHeader := r.Header.Get(requestHeaderOrigin)
	if !isValidOrigin(originHeader) {
		return invalidCorsRequestType
	} else if isSameOrigin(r.URL, originHeader) {
		return notCorsRequestType
	} else {
		switch r.Method {
		case http.MethodGet, http.MethodHead:
			return simpleCorsRequestType
		case http.MethodOptions:
			if _, found := r.Header[requestHeaderAccessControlRequestMethod]; found {
				if r.Header.Get(requestHeaderAccessControlRequestMethod) == "" {
					return invalidCorsRequestType
				}
				return preFlightCorsRequestType
			}
			return actualCorsRequestType
		case http.MethodPost:
			mediaType := getMediaType(r.Header.Get(requestHeaderContentType))
			if mediaType != "" {
				if mediaType == contentTypeValueTextPlain ||
					mediaType == contentTypeValueFormData ||
					mediaType == contentTypeValueFormUrlencodedData {
					return simpleCorsRequestType
				}
				return actualCorsRequestType
			}
		default:
			return actualCorsRequestType
		}
	}

	return invalidCorsRequestType
}

func isValidOrigin(origin string) bool {
	if origin == "" || strings.ContainsRune(origin, '%') {
		return false
	}

	if origin == "null" || strings.Index(origin, "file://") == 0 {
		return true
	}

	parse, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return parse.Scheme != ""
}

func isSameOrigin(reqUrl *url.URL, origin string) bool {
	if reqUrl.Scheme == "" || reqUrl.Host == "" {
		return false
	}

	scheme := strings.ToLower(reqUrl.Scheme)
	var sb strings.Builder
	sb.WriteString(scheme)
	sb.WriteString("://")
	sb.WriteString(reqUrl.Host)

	port := reqUrl.Port()
	if port == "" {
		if "https" == scheme || "wss" == scheme {
			port = "443"
		} else {
			port = "80"
		}
	}
	if sb.Len() == len(origin) {
		// origin and target can only be equal if both are using default ports
		if (("http" == scheme || "ws" == scheme) && port != "80") ||
			(("https" == scheme || "wss" == scheme) && port != "443") {
			return false
		}
	} else {
		sb.WriteString(":")
		sb.WriteString(port)
	}

	// the CORS spec states this check should be case-sensitive
	return sb.String() == origin
}

func getMediaType(contentType string) string {
	if contentType == "" {
		return contentType
	}

	contentType = strings.ToLower(contentType)
	firstSemiColonIndex := strings.IndexRune(contentType, ';')
	if firstSemiColonIndex >= 0 {
		contentType = contentType[0:firstSemiColonIndex]
	}
	return strings.TrimSpace(contentType)
}
