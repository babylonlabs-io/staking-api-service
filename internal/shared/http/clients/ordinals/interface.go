package ordinals

import (
	"context"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

type OrdinalsClient interface {
	GetBaseURL() string
	GetDefaultRequestTimeout() int
	GetHttpClient() *http.Client
	/*
		FetchUTXOInfos fetches UTXO information from the ordinal service
		The response from ordinal service shall contain all requested UTXOs and in
		the same order as requested
	*/
	FetchUTXOInfos(ctx context.Context, utxos []types.UTXOIdentifier) ([]OrdinalsOutputResponse, *types.Error)
}
