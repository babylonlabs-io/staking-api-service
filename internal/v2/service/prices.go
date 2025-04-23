package v2service

import (
	"context"
	"errors"
	"strings"

	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

func (s *V2Service) GetLatestPrices(ctx context.Context) (map[string]float64, *types.Error) {
	// it happens in case config doesn't contain values to initialize coinmarketcap client
	if s.clients.CoinMarketCap == nil {
		err := errors.New("coin market cap API is not configured")
		return nil, types.NewInternalServiceError(err)
	}

	btcPrice, err := s.sharedService.GetLatestBTCPrice(ctx)
	if err != nil {
		return nil, types.NewInternalServiceError(err)
	}

	babyPrice, err := s.sharedService.GetLatestBABYPrice(ctx)
	if err != nil {
		return nil, types.NewInternalServiceError(err)
	}

	// for now we get only btc prices
	btcSymbol := strings.ToUpper(dbmodel.SymbolBTC)
	babySymbol := strings.ToUpper(dbmodel.SymbolBABY)
	return map[string]float64{
		btcSymbol:  btcPrice,
		babySymbol: babyPrice,
	}, nil
}
