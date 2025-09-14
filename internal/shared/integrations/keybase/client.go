package keybase

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/client"
)

type Client struct {
	httpClient *http.Client
}

func (c *Client) GetBaseURL() string {
	return "https://keybase.io"
}

func (c *Client) GetDefaultRequestTimeout() int {
	return 5000 // 5 seconds
}

func (c *Client) GetHttpClient() *http.Client {
	return c.httpClient
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
	}
}

func (c *Client) GetLogoURL(ctx context.Context, identity string) (string, error) {
	type empty struct{}
	type lookupResponse struct {
		Status struct {
			Code int    `json:"code"`
			Name string `json:"name"`
		} `json:"status"`
		Them []struct {
			ID       string `json:"id"`
			Pictures struct {
				Primary struct {
					URL string `json:"url"`
				} `json:"primary"`
			} `json:"pictures"`
		} `json:"them"`
	}

	params := make(url.Values)
	params.Add("fields", "pictures")
	params.Add("username", "ds")
	params.Add("key_suffix", identity)
	query := params.Encode()

	const endpoint = "/_/api/1.0/user/lookup.json"
	path := endpoint + "?" + query

	opts := &client.HttpClientOptions{
		Path:         path,
		TemplatePath: endpoint,
	}

	resp, err := client.SendRequest[empty, lookupResponse](ctx, c, http.MethodGet, opts, nil)
	if err != nil {
		return "", err
	}

	if len(resp.Them) == 0 {
		return "", fmt.Errorf("no pictures found for %q", identity)
	}

	url := resp.Them[0].Pictures.Primary.URL
	if url == "" {
		return "", fmt.Errorf("empty picture url for %q (keybase changed response?)", identity)
	}

	return url, nil
}
