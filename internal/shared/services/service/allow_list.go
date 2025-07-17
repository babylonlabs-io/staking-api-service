package service

import "context"

func (s *Service) IsTxInAllowList(ctx context.Context, stakingTxHash string) (bool, error) {
	return s.DbClients.SharedDBClient.IsTxInAllowList(ctx, stakingTxHash)
}
