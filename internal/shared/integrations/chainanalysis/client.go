package chainanalysis

import (
	"errors"
	ch "github.com/0xFredZhang/chainalysis-go"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
)

type Interface interface {
	AssessAddress(address string) (any, error)
}

type Client struct {
	impl *ch.ClientImpl
}

func NewClient(apiKey string, host string) *Client {
	return &Client{
		impl: ch.NewClient(apiKey, host),
	}
}

func (c *Client) AssessAddress(address string) (any, error) {
	// todo change type to string
	resp, err := c.impl.EntityAddressRetrieve(address)
	if err != nil {
		metrics.RecordChainAnalysisCall(true)
		return nil, err
	}

	if resp.Message != "" {
		metrics.RecordChainAnalysisCall(true)
		return nil, errors.New(resp.Message)
	}

	metrics.RecordChainAnalysisCall(false)
	return resp.Risk, nil
}
