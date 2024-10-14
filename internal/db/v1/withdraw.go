package v1db

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/types"
	"github.com/babylonlabs-io/staking-api-service/internal/utils"
)

func (v1db *V1Database) TransitionToWithdrawnState(ctx context.Context, txHashHex string) error {
	err := v1db.transitionState(
		ctx, txHashHex, types.Withdrawn.ToString(),
		utils.QualifiedStatesToWithdraw(), nil,
	)
	if err != nil {
		return err
	}
	return nil
}
