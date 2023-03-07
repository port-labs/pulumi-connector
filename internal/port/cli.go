package port

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/go-resty/resty/v2"
)

type (
	AccessTokenResponse struct {
		Ok          bool   `json:"ok"`
		AccessToken string `json:"accessToken"`
		ExpiresIn   int64  `json:"expiresIn"`
		TokenType   string `json:"tokenType"`
	}
	Option func(*Client)
	Client struct {
		Client   *resty.Client
		ClientID string
	}
)

func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		Client: resty.New().
			SetBaseURL(baseURL).
			SetRetryCount(5).
			SetRetryWaitTime(300).
			// retry when create permission fails because scopes are created async-ly and sometimes (mainly in tests) the scope doesn't exist yet.
			AddRetryCondition(func(r *resty.Response, err error) bool {
				if err != nil {
					return true
				}
				if !strings.Contains(r.Request.URL, "/permissions") {
					return false
				}
				b := make(map[string]interface{})
				err = json.Unmarshal(r.Body(), &b)
				return err != nil || b["ok"] != true
			}),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) Authenticate(ctx context.Context, clientID, clientSecret string) (string, error) {
	url := "v1/auth/access_token"
	resp, err := c.Client.R().
		SetBody(map[string]interface{}{
			"clientId":     clientID,
			"clientSecret": clientSecret,
		}).
		SetContext(ctx).
		Post(url)
	if err != nil {
		return "", err
	}
	var tokenResp AccessTokenResponse
	err = json.Unmarshal(resp.Body(), &tokenResp)
	if err != nil {
		return "", err
	}
	c.Client.SetAuthToken(tokenResp.AccessToken)
	return tokenResp.AccessToken, nil
}

func WithHeader(key, val string) Option {
	return func(pc *Client) {
		pc.Client.SetHeader(key, val)
	}
}

func WithClientID(clientID string) Option {
	return func(pc *Client) {
		pc.ClientID = clientID
	}
}

func WithToken(token string) Option {
	return func(pc *Client) {
		pc.Client.SetAuthToken(token)
	}
}
