package v2service

import (
	"context"
	"math/rand"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/utils/datagen"
)

type GlobalParamsPublic struct {
	Babylon []types.BabylonParams `json:"babylon"`
	BTC     []types.BTCParams     `json:"btc"`
}

func (s *V2Service) GetGlobalParams(ctx context.Context) (GlobalParamsPublic, *types.Error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	babylonParams := datagen.GenerateRandomBabylonParams(r)
	btcParams := datagen.GenerateRandomBTCParams(r)
	return GlobalParamsPublic{
		Babylon: []types.BabylonParams{babylonParams},
		BTC:     []types.BTCParams{btcParams},
	}, nil
}
