package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ixtendio/gofre/auth"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var GitHubAuthorizeEndpoint = "https://github.com/login/oauth/authorize"
var GitHubAccessTokenEndpoint = "https://github.com/login/oauth/access_token"
var GitHubUserInfoEndpoint = "https://api.github.com/user"
var GitHubUserInfoProfileScopes = []string{"user"}

// GitHubProvider implements the OAUTH2 flow using GitHub
// For more information please visit https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps
type GitHubProvider struct {
	// The ClientId is the client ID you received from GitHub when you register your OAUTH app (REQUIRED)
	ClientId string

	// The ClientSecret is the client secret you received from GitHub when you register your OAUTH app (REQUIRED)
	ClientSecret string

	// Login suggests a specific account to use for signing in and authorizing the app (OPTIONAL)
	Login string

	// AllowSignup specifies whether or not unauthenticated users will be offered an option to sign up for GitHub during the OAuth flow
	AllowSignup bool

	// Scopes is a list of authorization requested scopes. Default read:user, user:email (OPTIONAL)
	Scopes []string
}

func (p GitHubProvider) Name() string {
	return "github"
}

func (p GitHubProvider) InitiateUrl(redirectUri string, state string, includeUserInfoProfileScope bool) string {
	scopes := p.Scopes
	if includeUserInfoProfileScope {
		scopes = append(scopes, GitHubUserInfoProfileScopes...)
	}
	urlValues := url.Values{
		"client_id":    {p.ClientId},
		"state":        {state},
		"allow_signup": {strconv.FormatBool(p.AllowSignup)},
		"redirect_uri": {redirectUri},
	}
	if p.Login != "" {
		urlValues.Add("login", p.Login)
	}
	if len(p.Scopes) > 0 {
		urlValues.Add("scope", strings.Join(scopes, " "))
	}
	return GitHubAuthorizeEndpoint + "?" + urlValues.Encode()
}

func (p GitHubProvider) FetchAccessToken(ctx context.Context, redirectUri string, authCode string) (AccessToken, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, GitHubAccessTokenEndpoint, strings.NewReader(url.Values{
		"client_id":     {p.ClientId},
		"client_secret": {p.ClientSecret},
		"code":          {authCode},
	}.Encode()))
	if err != nil {
		return AccessToken{}, fmt.Errorf("failed to create the request to get the OAUTH2 access token from: %s, err: %w", p.Name(), err)
	}

	req.Header.Set("Accept", "application/json")
	resp, err := HttpClient.Do(req)

	if err != nil {
		return AccessToken{}, fmt.Errorf("failed to fetch the OAUTH2 access token from: %s, err: %w", p.Name(), err)
	}
	defer resp.Body.Close()

	result := make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return AccessToken{}, fmt.Errorf("failed to decode the OAUTH2 access token response from: %s, err: %w", p.Name(), err)
	}

	if resp.StatusCode/100 == 2 {
		accessToken := AccessToken{
			AccessToken: result["access_token"].(string),
			TokenType:   result["token_type"].(string),
		}

		if scope, ok := result["scope"].(string); ok {
			accessToken.Scopes = strings.Split(scope, ",")
		}
		return accessToken, nil
	}

	if errMsg, found := result["error"]; found {
		return AccessToken{}, fmt.Errorf("failed to get the OAUTH2 access token response from: %s, err: %s", p.Name(), errMsg)
	}
	return AccessToken{}, fmt.Errorf("get the OAUTH2 access token response from: %s returned: %d", p.Name(), resp.StatusCode)

}

func (p GitHubProvider) FetchAuthenticatedUser(ctx context.Context, accessToken AccessToken) (auth.User, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, GitHubUserInfoEndpoint, nil)
	if err != nil {
		return auth.User{}, fmt.Errorf("failed to create the request to get the user details from: %s, err: %w", p.Name(), err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken.AccessToken)
	req.Header.Set("Accept", "application/json")
	resp, err := HttpClient.Do(req)

	if err != nil {
		return auth.User{}, fmt.Errorf("failed to fetch the user details from: %s, err: %w", p.Name(), err)
	}
	defer resp.Body.Close()

	result := make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return auth.User{}, fmt.Errorf("failed to decode the user details response from: %s, err: %w", p.Name(), err)
	}

	if resp.StatusCode/100 == 2 {
		var id string
		var name string
		if email, ok := result["email"].(string); ok {
			id = email
		} else {
			id = result["login"].(string)
		}
		if n, ok := result["name"].(string); ok {
			name = n
		}
		return auth.User{
			Id:               id,
			Name:             name,
			IdentityPlatform: p.Name(),
		}, nil
	}
	return auth.User{}, fmt.Errorf("get the user details response from: %s returned: %d", p.Name(), resp.StatusCode)
}
