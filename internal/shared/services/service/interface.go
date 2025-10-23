package service

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

type SharedServiceProvider interface {
	DoHealthCheck(ctx context.Context) error
	VerifyUTXOs(ctx context.Context, utxos []types.UTXOIdentifier, address string) ([]*SafeUTXOPublic, *types.Error)
	SaveUnprocessableMessages(ctx context.Context, messages string, receipt string) *types.Error
}
