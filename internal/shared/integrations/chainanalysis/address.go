package chainanalysis

import (
	"context"
	"errors"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/client"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	"net/http"
)

// empty is auxiliary type that is used for sending empty request body in the request
type empty struct{}

type AddressAssessment struct {
	Risk       string
	RiskReason *string
}

// not all fields presented here
type riskEntityResponse struct {
	Message string `json:"message"`

	Address    string `json:"address"`
	Risk       string `json:"risk"` // Severe, High, Medium, Low
	RiskReason string `json:"riskReason"`
}

func (c *Client) AssessAddress(ctx context.Context, address string) (*AddressAssessment, error) {
	resp, err := c.testMe(ctx, address)
	if err != nil {
		metrics.RecordChainAnalysisCall(true)
		return nil, err
	}

	if resp.Message != "" {
		metrics.RecordChainAnalysisCall(true)
		return nil, errors.New(resp.Message)
	}

	metrics.RecordChainAnalysisCall(false)

	var riskReason *string
	if resp.RiskReason != "" {
		riskReason = &resp.RiskReason
	}
	return &AddressAssessment{
		Risk:       resp.Risk,
		RiskReason: riskReason,
	}, nil
}

func (c *Client) testMe(ctx context.Context, address string) (*riskEntityResponse, error) {
	const endpoint = "/api/risk/v2/entities/"
	path := c.baseURL + endpoint + address

	opts := &client.HttpClientOptions{
		Path:         path,
		TemplatePath: endpoint,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Token":        c.apiKey,
		},
	}

	return client.SendRequest[empty, riskEntityResponse](
		ctx, c, http.MethodGet, opts, &empty{},
	)
}
