package ordinals

import (
	"context"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

//go:generate mockery --name=OrdinalsClient --output=../../../../../tests/mocks --outpkg=mocks --filename=mock_ordinal_client.go
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
