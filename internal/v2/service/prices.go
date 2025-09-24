package v2service

import (
	"context"
	"strings"

	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

func (s *V2Service) GetLatestPrices(ctx context.Context) (map[string]float64, *types.Error) {
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
