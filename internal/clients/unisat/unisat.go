package unisat

import (
	"context"
	"fmt"
	"net/http"

	baseclient "github.com/babylonchain/staking-api-service/internal/clients/base"
	"github.com/babylonchain/staking-api-service/internal/config"
	"github.com/babylonchain/staking-api-service/internal/types"
)

// Note: The JSON tags use camel case because this struct is used to
// represent the response from the Unisat endpoint, which uses camel case.
type UnisatInscriptions struct {
	InscriptionId     string `json:"inscriptionId"`
	InscriptionNumber uint32 `json:"inscriptionNumber"`
	IsBRC20           bool   `json:"isBRC20"`
	Moved             bool   `json:"moved"`
	Offset            uint32 `json:"offset"`
}

type UnisatUTXO struct {
	TxId         string                `json:"txid"`
	Vout         uint32                `json:"vout"`
	Inscriptions []*UnisatInscriptions `json:"inscriptions"`
}

type UnisatResponseData struct {
	Cursor uint32        `json:"cursor"`
	Total  uint32        `json:"total"`
	Utxo   []*UnisatUTXO `json:"utxo"`
}

// Refer to https://open-api.unisat.io/swagger.html
type UnisatResponse struct {
	Code int                `json:"code"`
	Msg  string             `json:"msg"`
	Data UnisatResponseData `json:"data"`
}

type UnisatClient struct {
	config        *config.UnisatConfig
	httpClient    *http.Client
	defaultHeader map[string]string
}

func NewUnisatClient(config *config.UnisatConfig) *UnisatClient {
	// Client is disabled if config is nil
	if config == nil {
		return nil
	}
	httpClient := &http.Client{}
	defaultHeader := map[string]string{
		"Accept":        "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", config.ApiToken),
	}
	return &UnisatClient{
		config,
		httpClient,
		defaultHeader,
	}
}

// Necessary for the BaseClient interface
func (c *UnisatClient) GetBaseURL() string {
	return fmt.Sprintf("%s", c.config.Host)
}

func (c *UnisatClient) GetDefaultRequestTimeout() int {
	return c.config.Timeout
}

func (c *UnisatClient) GetHttpClient() *http.Client {
	return c.httpClient
}

func (c *UnisatClient) FetchInscriptionsUTXOsByAddress(
	ctx context.Context, address string, cursor uint32,
) ([]*UnisatUTXO, *types.Error) {
	path := fmt.Sprintf(
		"/v1/indexer/address/%s/inscription-utxo-data?cursor=%d&size=%d",
		address, cursor, c.config.Limit,
	)
	opts := &baseclient.BaseClientOptions{
		Path:         path,
		TemplatePath: "/v1/indexer/address/{address}/inscription-utxo-data",
		Headers:      c.defaultHeader,
	}

	resp, err := baseclient.SendRequest[any, UnisatResponse](
		ctx, c, http.MethodGet, opts, nil,
	)
	if err != nil {
		return nil, err
	}

	return resp.Data.Utxo, nil
}
