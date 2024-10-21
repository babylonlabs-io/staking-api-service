package service

import (
	"context"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/rs/zerolog/log"
)

type SafeUTXOPublic struct {
	TxId        string `json:"txid"`
	Vout        uint32 `json:"vout"`
	Inscription bool   `json:"inscription"`
}

func (s *Service) VerifyUTXOs(
	ctx context.Context, utxos []types.UTXOIdentifier, address string,
) ([]*SafeUTXOPublic, *types.Error) {
	result, err := s.verifyViaOrdinalService(ctx, utxos)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg(
			"failed to verify ordinals via ordinals service",
		)
		return nil, err
	}
	return result, nil
}

func (s *Service) verifyViaOrdinalService(
	ctx context.Context, utxos []types.UTXOIdentifier,
) ([]*SafeUTXOPublic, *types.Error) {
	var results []*SafeUTXOPublic

	outputs, err := s.Clients.Ordinals.FetchUTXOInfos(ctx, utxos)
	if err != nil {
		return nil, err
	}

	for index, output := range outputs {
		// Check the order of the response is the same as the request
		if output.Transaction != utxos[index].Txid {
			return nil, types.NewErrorWithMsg(
				http.StatusInternalServerError,
				types.InternalServiceError,
				"ordinal service response order does not match the request",
			)
		}
		hasInscription := false

		// Check if Runes is not an empty JSON object
		if len(output.Runes) > 0 && string(output.Runes) != "{}" {
			hasInscription = true
		} else if len(output.Inscriptions) > 0 { // Check if Inscriptions is not empty
			hasInscription = true
		}
		results = append(results, &SafeUTXOPublic{
			TxId:        output.Transaction,
			Vout:        utxos[index].Vout,
			Inscription: hasInscription,
		})
	}

	return results, nil
}
