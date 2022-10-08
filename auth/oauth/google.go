package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ixtendio/gow/auth"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var GoogleAuthorizeEndpoint = "https://accounts.google.com/o/oauth2/v2/auth"
var GoogleAccessTokenEndpoint = "https://oauth2.googleapis.com/token"
var GoogleUserInfoEndpoint = "https://www.googleapis.com/oauth2/v1/userinfo"
var GoogleUserInfoScopes = []string{"openid", "https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"}

// GoogleProvider implements the OAUTH2 flow using Google
type GoogleProvider struct {
	// The ClientId is the client ID you received from Google when you register your OAUTH app (REQUIRED)
	ClientId string

	// The ClientSecret is the client secret you received from Google when you register your OAUTH app (REQUIRED)
	ClientSecret string

	// Scopes is a list scopes that identify the resources that your application could access on the user's behalf (REQUIRED)
	Scopes []string

	// The AccessTypeOffline indicates whether your application can refresh access tokens when the user is not present at the browser
	AccessTypeOffline bool

	// The IncludeGrantedScopes enables the application to use incremental authorization to request access to additional scopes in context (see: https://developers.google.com/identity/protocols/oauth2/web-server#incrementalAuth)
	IncludeGrantedScopes bool

	// LoginHint provides a hint to the Google Authentication Server about which user is trying to authenticate (OPTIONAL)
	LoginHint string

	// Prompts represents and array of case-sensitive list of prompts to present the user. Possible values are: (none, consent, select_account) (OPTIONAL) (see:https://developers.google.com/identity/protocols/oauth2/openid-connect#re-consent)
	Prompts []string
}

func (p GoogleProvider) Name() string {
	return "google"
}

func (p GoogleProvider) InitiateUrl(redirectUri string, state string, includeUserInfoProfileScope bool) string {
	scopes := p.Scopes
	if includeUserInfoProfileScope {
		//openid email profile
		scopes = append(scopes, GoogleUserInfoScopes...)
	}
	urlValues := url.Values{
		"client_id":              {p.ClientId},
		"redirect_uri":           {redirectUri},
		"response_type":          {"code"},
		"scope":                  {strings.Join(scopes, " ")},
		"state":                  {state},
		"include_granted_scopes": {strconv.FormatBool(p.IncludeGrantedScopes)},
	}

	if p.AccessTypeOffline {
		urlValues.Add("access_type", "offline")
	}
	if p.LoginHint != "" {
		urlValues.Add("login_hint", p.LoginHint)
	}
	if len(p.Prompts) > 0 {
		urlValues.Add("prompt", strings.Join(p.Prompts, " "))
	}
	return GoogleAuthorizeEndpoint + "?" + urlValues.Encode()
}

func (p GoogleProvider) FetchAccessToken(ctx context.Context, redirectUri string, authCode string) (AccessToken, error) {
	reqPayload := map[string]string{
		"client_id":     p.ClientId,
		"client_secret": p.ClientSecret,
		"code":          authCode,
		"grant_type":    "authorization_code",
		"redirect_uri":  redirectUri,
	}
	reqPayloadData, err := json.Marshal(reqPayload)
	if err != nil {
		return AccessToken{}, fmt.Errorf("failed to encode as JSON the payload to request the OAUTH2 access token from: %s, err: %w", p.Name(), err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, GoogleAccessTokenEndpoint, bytes.NewReader(reqPayloadData))
	if err != nil {
		return AccessToken{}, fmt.Errorf("failed to create the request to get the OAUTH2 access token from: %s, err: %w", p.Name(), err)
	}

	req.Header.Set("Content-Type", "application/json")
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

		if refreshToken, ok := result["refresh_token"].(string); ok {
			accessToken.RefreshToken = refreshToken
		}
		if expiresIn, ok := result["expires_in"].(float64); ok {
			accessToken.ExpiresInSeconds = int(expiresIn)
		}
		if scope, ok := result["scope"].(string); ok {
			accessToken.Scopes = strings.Split(scope, " ")
		}

		return accessToken, nil
	}

	if errObj, ok := result["error"].(map[string]interface{}); ok {
		if errMsg, found := errObj["message"]; found {
			return AccessToken{}, fmt.Errorf("failed to get the OAUTH2 access token response from: %s, err: %s", p.Name(), errMsg)
		}
	}
	return AccessToken{}, fmt.Errorf("get the OAUTH2 access token response from: %s returned: %d", p.Name(), resp.StatusCode)
}

func (p GoogleProvider) FetchAuthenticatedUser(ctx context.Context, accessToken AccessToken) (auth.User, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, GoogleUserInfoEndpoint, nil)
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
			id = result["id"].(string)
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
