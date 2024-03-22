package cf

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"time"
)

// makeAuthenticationRequest is a helper function to make an authentication request.
// it fills some default parameters for authentication requests.
func (cfg *CloudFoundryConfig) makeAuthenticationRequest(authParams map[string]string) (*resty.Response, error) {
	return resty.New().R().
		EnableTrace().
		SetQueryParams(authParams).
		SetBasicAuth(cfg.OAuthClientID, cfg.OAuthClientSecret).
		SetResult(new(AuthTokenInfo)).
		Post(fmt.Sprintf("%s/oauth/token", cfg.AuthEndpoint))
}

// NewClient returns a new request httpClient which manages the authToken and refreshes it if necessary
func (cfg *CloudFoundryConfig) NewClient() (*CloudFoundryClient, error) {
	resp, err := cfg.makeAuthenticationRequest(map[string]string{
		"grant_type": "password",
		"scope":      "",
		"username":   cfg.Username,
		"password":   cfg.Password,
	})
	if err != nil {
		return nil, err
	}

	token := resp.Result().(*AuthTokenInfo)
	return &CloudFoundryClient{
		authToken:    token,
		lastAuthTime: time.Now(),
		config:       cfg,
		httpClient:   resty.New(),
	}, nil
}

// TokenIsExpired returns true if the authToken is expired
func (req *CloudFoundryClient) TokenIsExpired() bool {
	return req.lastAuthTime.
		Add(time.Duration(req.authToken.ExpiresIn)*time.Second - TokenExpirySafetyMargin).
		Before(time.Now())
}

// RefreshToken tries to refresh the authToken
func (req *CloudFoundryClient) RefreshToken() error {
	resp, err := req.config.makeAuthenticationRequest(map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": req.authToken.RefreshToken,
	})
	if err != nil {
		return err
	}

	// Update the authentication token and timestamp based on the response
	req.authToken = resp.Result().(*AuthTokenInfo)
	req.lastAuthTime = time.Now()
	return nil
}
