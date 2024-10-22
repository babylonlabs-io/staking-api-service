package ordinals

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/client"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

type OrdinalsOutputResponse struct {
	Transaction  string          `json:"transaction"` // same as Txid
	Inscriptions []string        `json:"inscriptions"`
	Runes        json.RawMessage `json:"runes"`
}

type Ordinals struct {
	config         *config.OrdinalsConfig
	defaultHeaders map[string]string
	httpClient     *http.Client
}

func New(config *config.OrdinalsConfig) *Ordinals {
	// Client is disabled if config is nil
	if config == nil {
		return nil
	}
	httpClient := &http.Client{}
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}
	return &Ordinals{
		config,
		headers,
		httpClient,
	}
}

// Necessary for the BaseClient interface
func (c *Ordinals) GetBaseURL() string {
	return fmt.Sprintf("%s:%s", c.config.Host, c.config.Port)
}

func (c *Ordinals) GetDefaultRequestTimeout() int {
	return c.config.Timeout
}

func (c *Ordinals) GetHttpClient() *http.Client {
	return c.httpClient
}

func (c *Ordinals) FetchUTXOInfos(
	ctx context.Context, utxos []types.UTXOIdentifier,
) ([]OrdinalsOutputResponse, *types.Error) {
	path := "/outputs"
	opts := &client.HttpClientOptions{
		Path:         path,
		TemplatePath: path,
		Headers:      c.defaultHeaders,
	}

	var txHashVouts []string
	for _, utxo := range utxos {
		txHashVouts = append(txHashVouts, fmt.Sprintf("%s:%d", utxo.Txid, utxo.Vout))
	}

	outputsResponse, err := client.SendRequest[[]string, []OrdinalsOutputResponse](
		ctx, c, http.MethodPost, opts, &txHashVouts,
	)
	if err != nil {
		return nil, err
	}
	outputs := *outputsResponse

	// The response from ordinal service shall contain all requested UTXOs and in
	// the same order
	for i, utxo := range utxos {
		if outputs[i].Transaction != utxo.Txid {
			return nil, types.NewErrorWithMsg(
				http.StatusInternalServerError,
				types.InternalServiceError,
				"response does not contain all requested UTXOs or in the wrong order",
			)
		}
	}

	return outputs, nil
}
