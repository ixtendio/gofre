package middleware

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/ixtendio/gow/cache"
	"github.com/ixtendio/gow/errors"
	"github.com/ixtendio/gow/request"
	"github.com/ixtendio/gow/response"
	"github.com/ixtendio/gow/router"
	"net/http"
	"strings"
	"time"
)

type csrfCtxKey int

var CSRFExpirationTime = 1 * time.Hour

const CSRFNonceKey csrfCtxKey = 1
const CSRFNonceRequestParamName = "_csrf"
const CSRFRestNonceHeaderName = "X-CSRF-Token"

// CSRFPrevention provides basic CSRF protection for a web application
func CSRFPrevention(nonceCache cache.Cache) Middleware {
	return CSRFPreventionWithCustomParamAndHeaderName(nonceCache, CSRFNonceRequestParamName, CSRFRestNonceHeaderName)
}

// CSRFPreventionWithCustomParamAndHeaderName provides basic CSRF protection for a web application using a custom form param name and header name
func CSRFPreventionWithCustomParamAndHeaderName(nonceCache cache.Cache, csrfNonceRequestParamName string, csrfNonceRequestHeaderName string) Middleware {
	return func(handler router.Handler) router.Handler {
		return func(ctx context.Context, req *request.HttpRequest) (response.HttpResponse, error) {
			skipNonceCheck := req.R.Method == http.MethodGet ||
				req.R.Method == http.MethodHead ||
				req.R.Method == http.MethodTrace ||
				req.R.Method == http.MethodOptions
			if !skipNonceCheck {
				previousNonce := req.R.Header.Get(csrfNonceRequestHeaderName)
				if previousNonce == "" {
					previousNonce = req.R.Form.Get(csrfNonceRequestParamName)
				}
				if previousNonce == "" || !nonceCache.Contains(previousNonce) {
					return nil, errors.ErrDenied
				}
				nonceCache.Remove(previousNonce)
			}

			newNonce, err := generateNonce()
			if err != nil {
				return nil, err
			}
			if err := nonceCache.Add(newNonce, CSRFExpirationTime); err != nil {
				return nil, err
			}
			ctx = context.WithValue(ctx, CSRFNonceKey, newNonce)
			return handler(ctx, req)
		}
	}
}

func generateNonce() (string, error) {
	var sb strings.Builder

	randBytes := make([]byte, 16)
	_, err := rand.Read(randBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate nonce, err: %w", err)
	}

	for _, b := range randBytes {
		b1 := (b & 0xf0) >> 4
		b2 := b & 0x0f
		if b1 < 10 {
			sb.WriteByte('0' + b1)
		} else {
			sb.WriteByte('A' + (b1 - 10))
		}
		if b2 < 10 {
			sb.WriteByte('0' + b2)
		} else {
			sb.WriteByte('A' + (b2 - 10))
		}
	}

	return sb.String(), nil
}
