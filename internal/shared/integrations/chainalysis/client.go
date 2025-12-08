package chainalysis

import (
	"context"
	"net/http"
	"time"
)

const defaultTimeout = 5 * time.Second

type Interface interface {
	AssessAddress(ctx context.Context, address string) (*AddressAssessment, error)
}

type Client struct {
	httpClient *http.Client

	apiKey  string
	baseURL string
}

func (c *Client) GetBaseURL() string {
	return c.baseURL
}

func (c *Client) GetDefaultRequestTimeout() time.Duration {
	return defaultTimeout
}

func (c *Client) GetHttpClient() *http.Client {
	return c.httpClient
}

func NewClient(apiKey string, baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{},
		apiKey:     apiKey,
		baseURL:    baseURL,
	}
}
