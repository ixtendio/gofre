package middleware

import (
	"context"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/router/path"

	"github.com/ixtendio/gofre/response"
	"strconv"
	"strings"
)

type XFrameOption int

const (
	stsHeaderName                                    = "Strict-Transport-Security"
	antiClickJackingHeaderName                       = "X-Frame-Options"
	blockContentTypeSniffingHeaderName               = "X-Content-Type-Options"
	blockContentTypeSniffingHeaderValue              = "nosniff"
	xssProtectionHeaderName                          = "X-Xss-Protection"
	xssProtectionHeaderValue                         = "1; mode=block"
	XFrameOptionDeny                    XFrameOption = iota
	XFrameOptionSameOrigin
	XFrameOptionAllowFrom
)

type ShStrictTransportSecurityConfig struct {
	Enabled           bool
	MaxAgeSeconds     int
	IncludeSubDomains bool
	Preload           bool
	headerValue       string
}

// Build STS header value
func (c *ShStrictTransportSecurityConfig) getHeaderValue() string {
	if c.headerValue == "" {
		var sb strings.Builder
		sb.WriteString("max-age=")
		sb.WriteString(strconv.Itoa(c.MaxAgeSeconds))
		if c.IncludeSubDomains {
			sb.WriteString(";includeSubDomains")
		}
		if c.Preload {
			sb.WriteString(";preload")
		}
		c.headerValue = sb.String()
	}
	return c.headerValue
}

type ShClickJackingConfig struct {
	Enabled                 bool
	XFrameOption            XFrameOption
	XFrameOptionHeaderValue string
	XFrameAllowFromUri      string
	headerValue             string
}

// Anti click-jacking
func (c *ShClickJackingConfig) getAntiClickJackingHeaderValue() string {
	if c.headerValue == "" {
		var sb strings.Builder
		sb.WriteString(c.XFrameOptionHeaderValue)
		if c.XFrameOption == XFrameOptionAllowFrom {
			sb.WriteRune(' ')
			sb.WriteString(c.XFrameAllowFromUri)
		}
		c.headerValue = sb.String()
	}
	return c.headerValue
}

type SecurityHeadersConfig struct {
	STS                         ShStrictTransportSecurityConfig
	ClickJacking                ShClickJackingConfig
	BlockContentSniffingEnabled bool
	XSSProtectionEnabled        bool
}

// SecurityHeaders provides some security HTTP headers to the response
func SecurityHeaders(config SecurityHeadersConfig) Middleware {
	return func(handler handler.Handler) handler.Handler {
		return func(ctx context.Context, mc path.MatchingContext) (response.HttpResponse, error) {
			httpResponse, err := handler(ctx, mc)
			if err != nil {
				return nil, err
			}

			// HSTS
			isRequestSecure := mc.R.TLS != nil || strings.ToLower(mc.R.URL.Scheme) == "https"
			if config.STS.Enabled && isRequestSecure {
				httpResponse.Headers().Set(stsHeaderName, config.STS.getHeaderValue())
			}

			// anti click-jacking
			if config.ClickJacking.Enabled {
				httpResponse.Headers().Set(antiClickJackingHeaderName, config.ClickJacking.getAntiClickJackingHeaderValue())
			}

			// Block content type sniffing
			if config.BlockContentSniffingEnabled {
				httpResponse.Headers().Set(blockContentTypeSniffingHeaderName,
					blockContentTypeSniffingHeaderValue)
			}

			// cross-site scripting filter protection
			if config.XSSProtectionEnabled {
				httpResponse.Headers().Set(xssProtectionHeaderName, xssProtectionHeaderValue)
			}

			return httpResponse, nil
		}
	}
}
