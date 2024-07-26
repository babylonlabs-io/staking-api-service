package services

import (
	"context"
	"fmt"

	"github.com/babylonchain/staking-api-service/internal/clients/unisat"
	"github.com/babylonchain/staking-api-service/internal/types"
	"github.com/rs/zerolog/log"
)

type SafeUTXOPublic struct {
	TxId        string `json:"txid"`
	Vout        uint32 `json:"vout"`
	Inscription bool   `json:"inscription"`
}

func (s *Services) VerifyUTXOs(
	ctx context.Context, utxos []types.UTXOIdentifier, address string,
) ([]*SafeUTXOPublic, *types.Error) {
	result, err := s.verifyViaOrdinalService(ctx, utxos)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg(
			"failed to verify ordinals via ordinals service",
		)
		unisatResult, err := s.verifyViaUnisatService(ctx, address, utxos)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg(
				"failed to verify ordinals via unisat service",
			)
			return nil, err
		}
		return unisatResult, nil
	}
	return result, nil
}

func (s *Services) verifyViaOrdinalService(
	ctx context.Context, utxos []types.UTXOIdentifier,
) ([]*SafeUTXOPublic, *types.Error) {
	var results []*SafeUTXOPublic

	outputs, err := s.Clients.Ordinals.FetchUTXOInfos(ctx, utxos)
	if err != nil {
		return nil, err
	}

	for index, output := range outputs {
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

func (s *Services) verifyViaUnisatService(
	ctx context.Context, address string, utxos []types.UTXOIdentifier,
) ([]*SafeUTXOPublic, *types.Error) {
	cursor := uint32(0)
	var inscriptionsUtxos []*unisat.UnisatUTXO
	limit := s.cfg.Assets.Unisat.Limit

	for {
		inscriptions, err := s.Clients.Unisat.FetchInscriptionsUtxosByAddress(
			ctx, address, cursor,
		)
		if err != nil {
			return nil, err
		}
		// Append the fetched utxos to the list
		inscriptionsUtxos = append(inscriptionsUtxos, inscriptions...)
		// Stop fetching if the total number of utxos is less than the limit
		if uint32(len(inscriptions)) < limit {
			break
		}
		// update the cursor for the next fetch
		cursor += limit
	}

	// turn inscriptionsUtxos into a map for easier lookup
	inscriptionsUtxosMap := make(map[string][]*unisat.UnisatInscriptions)
	for _, inscriptionsUtxo := range inscriptionsUtxos {
		key := fmt.Sprintf("%s:%d", inscriptionsUtxo.TxId, inscriptionsUtxo.Vout)
		inscriptionsUtxosMap[key] = inscriptionsUtxo.Inscriptions
	}

	var results []*SafeUTXOPublic
	for _, utxo := range utxos {
		key := fmt.Sprintf("%s:%d", utxo.Txid, utxo.Vout)
		inscriptions, ok := inscriptionsUtxosMap[key]
		results = append(results, &SafeUTXOPublic{
			TxId:        utxo.Txid,
			Vout:        utxo.Vout,
			Inscription: ok && len(inscriptions) > 0,
		})
	}
	return results, nil
}
