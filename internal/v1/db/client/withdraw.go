package v1dbclient

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/utils"
)

func (v1dbclient *V1Database) TransitionToWithdrawnState(ctx context.Context, txHashHex string) error {
	err := v1dbclient.transitionState(
		ctx, txHashHex, types.Withdrawn.ToString(),
		utils.QualifiedStatesToWithdraw(), nil,
	)
	if err != nil {
		return err
	}
	return nil
}
