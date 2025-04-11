package chainalysis

import (
	"context"
	"net/http"
	"strings"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/client"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
)

// empty is auxiliary type that is used for sending empty request body in the request
type empty struct{}

type AddressAssessment struct {
	Risk       string
	RiskReason *string
}

type Cluster struct {
	Name string `json:"string"`
}

// not all fields presented here
type riskEntityResponse struct {
	Message string `json:"message"`

	Address    string   `json:"address"`
	Risk       string   `json:"risk"` // Severe, High, Medium, Low
	RiskReason string   `json:"riskReason"`
	Cluster    *Cluster `json:"cluster"`
}

func (c *Client) AssessAddress(ctx context.Context, address string) (*AddressAssessment, error) {
	resp, err := c.doAccessAddress(ctx, address)
	if err != nil {
		metrics.RecordChainAnalysisCall(true)
		return nil, err
	}

	metrics.RecordChainAnalysisCall(false)
	metrics.RecordAssessAddress(resp.Risk)

	// based on discussion with chainalysis we allow xverse.app addresses, but only
	// those that don't contain "Identified" in the risk reason
	isXverse := resp.Cluster != nil && resp.Cluster.Name == "Xverse.app"
	reasonContainsIdentified := strings.Contains(resp.RiskReason, "Identified")
	if isXverse && !reasonContainsIdentified {
		return &AddressAssessment{
			Risk: "Low",
		}, nil
	}

	var riskReason *string
	if resp.RiskReason != "" {
		riskReason = &resp.RiskReason
	}
	return &AddressAssessment{
		Risk:       resp.Risk,
		RiskReason: riskReason,
	}, nil
}

func (c *Client) doAccessAddress(ctx context.Context, address string) (*riskEntityResponse, error) {
	const endpoint = "/api/risk/v2/entities/"
	path := endpoint + address

	opts := &client.HttpClientOptions{
		Path:         path,
		TemplatePath: endpoint,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Token":        c.apiKey,
		},
	}

	resp, err := client.SendRequest[empty, riskEntityResponse](
		ctx, c, http.MethodGet, opts, &empty{},
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
