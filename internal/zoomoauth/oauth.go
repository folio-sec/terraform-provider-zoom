package zoomoauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	ApiURL       string `json:"api_url"`
}

type Client struct {
	httpclient HttpRequestDoer
}

type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

func NewClient(opts ...ClientOption) (*Client, error) {
	client := Client{
		httpclient: http.DefaultClient,
	}
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.httpclient = doer
		return nil
	}
}

func (c *Client) GetAccessToken(ctx context.Context, accountID string, clientID string, clientSecret string) (*TokenResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://zoom.us/oauth/token",
		strings.NewReader(url.Values{
			"grant_type": {"account_credentials"},
			"account_id": {accountID},
		}.Encode()),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)
	req.Header.Set(
		"Authorization",
		fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret)))),
	)

	res, err := c.httpclient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth token: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			// TODO
		}
	}(res.Body)

	resp, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read, io.ReadAll(): %w", err)
	}

	var tokenResponse TokenResponse
	err = json.Unmarshal(resp, &tokenResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse json payload, json.Unmarshal(): %w", err)
	}

	return &tokenResponse, nil
}
