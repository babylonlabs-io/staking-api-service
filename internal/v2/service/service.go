package v2service

import (
	"context"
	"fmt"
	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services/service"
	"github.com/davecgh/go-spew/spew"
	"slices"
)

type V2Service struct {
	dbClients     *dbclients.DbClients
	clients       *clients.Clients
	cfg           *config.Config
	sharedService *service.Service
}

func (v *V2Service) SelfCheck(ctx context.Context) error {
	items, err := v.dbClients.IndexerDBClient.GetAllDelegations(ctx)
	if err != nil {
		return err
	}

	for i, item := range items {
		pkHEX := item.StakerBtcPkHex

		fmt.Printf("Processing %d item\n", i)
		if slices.Contains([]string{
			"f33a3851632079b01be26360bba9f7f49286406d2636a25565f8d83db6ac2a32",
			"59b0de01b3dd4fe04438153a0da76a40da9674fdb2034fcc8b1c33a5b09ab1e1",
			"c35a85355bf690c80ac54c3e8e9916b04d464a4b37b6f99091e046c65f05c069",
			"19ebd9a0d8fd611c4e406512a9df0073a135855e71f11f488cbda08d5b46b8c6",
		}, pkHEX) {
			fmt.Printf("Skipping %s\n", pkHEX)
			continue
		}
		err = v.check(ctx, pkHEX)
		if err != nil {
			if db.IsNotFoundError(err) {
				fmt.Printf("Record for %v is not found\n", pkHEX)
				continue
			}
			return err
		}
	}

	panic("All good")
}

func (v *V2Service) check(ctx context.Context, stakerPKHex string) error {
	states := []indexertypes.DelegationState{
		indexertypes.StateActive,
		indexertypes.StateUnbonding,
		indexertypes.StateWithdrawable,
		indexertypes.StateSlashed, // do we need slashed here ?
	}
	delegations, err := v.dbClients.IndexerDBClient.GetDelegationsInStates(ctx, stakerPKHex, states)
	if err != nil {
		return err
	}

	var stats StakerStatsPublic
	stats.StakerPkHex = stakerPKHex

	for _, delegation := range delegations {
		amount := int64(delegation.StakingAmount)

		switch delegation.State {
		case indexertypes.StateActive:
			stats.ActiveTvl += amount
			stats.ActiveDelegations++
		case indexertypes.StateUnbonding:
			stats.UnbondingTvl += amount
			stats.UnbondingDelegations++
		case indexertypes.StateWithdrawable:
			stats.WithdrawableTvl += amount
			stats.WithdrawableDelegations++
		}
	}

	newCtx := context.Background()
	stakerStats, err := v.dbClients.V2DBClient.GetStakerStats(newCtx, stakerPKHex)
	if err != nil {
		return err
	}

	oldStats := StakerStatsPublic{
		StakerPkHex:             stakerStats.StakerPkHex,
		ActiveTvl:               stakerStats.ActiveTvl,
		ActiveDelegations:       stakerStats.ActiveDelegations,
		UnbondingTvl:            stakerStats.UnbondingTvl,
		UnbondingDelegations:    stakerStats.UnbondingDelegations,
		WithdrawableTvl:         stakerStats.WithdrawableTvl,
		WithdrawableDelegations: stakerStats.WithdrawableDelegations,
	}
	if oldStats != stats {
		spew.Dump("OLD", oldStats)
		spew.Dump("NEW", stats)
		return fmt.Errorf("staker pk %v has different stats", stakerStats.StakerPkHex)
	}

	return nil
}

func New(sharedService *service.Service) (*V2Service, error) {
	return &V2Service{
		dbClients:     sharedService.DbClients,
		clients:       sharedService.Clients,
		cfg:           sharedService.Cfg,
		sharedService: sharedService,
	}, nil
}
