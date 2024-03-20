package btp

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"time"
)

// makeAuthRequest is a helper function to make an authentication request.
// it fills some default parameters for authentication requests.
func (c *Config) makeAuthRequest(params map[string]string) (*resty.Response, error) {
	return resty.New().R().
		EnableTrace().
		SetQueryParams(params).
		SetBasicAuth(c.OAuthClientID, c.OAuthClientSecret).
		SetResult(new(TokenResponse)).
		Post(fmt.Sprintf("%s/oauth/token", c.AuthEndpoint))
}

// GetRequester returns a new requester which manages the token and refreshes it if necessary
func (c *Config) GetRequester() (*Requester, error) {
	resp, err := c.makeAuthRequest(map[string]string{
		"grant_type": "password",
		"scope":      "",
		"username":   c.Username,
		"password":   c.Password,
	})

	if err != nil {
		return nil, err
	}

	token := resp.Result().(*TokenResponse)
	fmt.Println("Got token:", token.AccessToken)
	return &Requester{
		token:  token,
		time:   time.Now(),
		config: c,
	}, nil
}

// IsExpired returns true if the token is expired
func (req *Requester) IsExpired() bool {
	return req.time.
		Add(time.Duration(req.token.ExpiresIn)*time.Second - TokenTimeJitter).
		Before(time.Now())
}

// Refresh tries to refresh the token
func (req *Requester) Refresh() error {
	resp, err := req.config.makeAuthRequest(map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": req.token.RefreshToken,
	})
	if err != nil {
		return err
	}

	// update the token and time from the response
	req.token = resp.Result().(*TokenResponse)
	req.time = time.Now()
	return nil
}
