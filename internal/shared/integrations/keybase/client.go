package keybase

import (
	"context"
	"fmt"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/client"
	"net/http"
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

	const endpoint = "/_/api/1.0/user/lookup.json"
	path := endpoint + fmt.Sprintf("?key_suffix=%s&fields=pictures", identity)

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

	return resp.Them[0].Pictures.Primary.URL, nil
}
