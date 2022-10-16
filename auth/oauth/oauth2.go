package oauth

import (
	"context"
	"github.com/ixtendio/gofre/auth"
	"github.com/ixtendio/gofre/cache"
	"net"
	"net/http"
	"time"
)

// The HttpClient is an optimized http client used to exchange an authorization code for an AccessToken (OPTIONAL)
var HttpClient = &http.Client{
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
	},
	Timeout: 10 * time.Second,
}

type ctxKey int

// AccessTokenCtxKey is used to pass the AccessToken to the request context.Context
const AccessTokenCtxKey ctxKey = 1

// GetAccessTokenFromContext returns the OAUTH2 AccessToken from the request context.Context
func GetAccessTokenFromContext(ctx context.Context) AccessToken {
	if at, ok := ctx.Value(AccessTokenCtxKey).(AccessToken); ok {
		return at
	}
	return AccessToken{}
}

// A CacheConfig encapsulate the cache related values
type CacheConfig struct {
	// Cache should be an instance of a distributed cache.Cache implementation
	Cache cache.Cache
	// KeyExpirationTime specifies the expiration time for the keys from the cache
	KeyExpirationTime time.Duration
}

// A Config encapsulate the required information for the OAUTH2 flow
type Config struct {
	// WebsiteUrl the website domain including the https scheme without trailing slash (example: https://www.mydomain.com) (REQUIRED)
	WebsiteUrl string
	// If FetchUserDetails is true, the user details (id, name, emails) are requested
	FetchUserDetails bool
	// Providers is a list of OAUTH2 supported providers (REQUIRED)
	Providers []Provider
	// CacheConfig (OPTIONAL)
	CacheConfig
}

// GetProviderByName returns a Provider with the specified name, otherwise nil
func (c Config) GetProviderByName(providerName string) Provider {
	for _, p := range c.Providers {
		if providerName == p.Name() {
			return p
		}
	}
	return nil
}

// An AccessToken includes the fields returned by the OAUTH2 provider after the exchange of the authorization code
type AccessToken struct {
	// AccessToken the token that the application should send it to authorize a Provider request (REQUIRED)
	AccessToken string
	// TokenType the type of token returned (for example: Bearer) (REQUIRED)
	TokenType string
	// Scopes the scopes of access granted by the AccessToken (case-sensitive strings) (REQUIRED)
	Scopes []string
	// RefreshToken a token that you can use to obtain a new access token. Refresh tokens are valid until the user revokes access. (OPTIONAL)
	RefreshToken string
	// ExpiresInSeconds the remaining lifetime of the access token in seconds (OPTIONAL)
	ExpiresInSeconds int
}

func (t AccessToken) IsEmpty() bool {
	return len(t.AccessToken) == 0
}

// The Provider interface defines the methods for an OAUTH2 provider
type Provider interface {
	// Name returns the name of the OAUTH2 provider (e.g. github, google, facebook, twitter, linkedin)
	Name() string

	// InitiateUrl returns the initiate URL for the OAUTH2 flow
	// The state parameter should be an unguessable random string that will be used to protect against cross-site request forgery attacks
	InitiateUrl(redirectUri string, state string, includeUserInfoProfileScope bool) string

	// FetchAccessToken exchange the authorization code for the access token
	FetchAccessToken(ctx context.Context, redirectUri string, authCode string) (AccessToken, error)

	// FetchAuthenticatedUser retrieves the authenticated user details
	FetchAuthenticatedUser(ctx context.Context, accessToken AccessToken) (auth.User, error)
}
