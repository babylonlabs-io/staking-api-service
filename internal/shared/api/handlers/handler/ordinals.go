package handler

import (
	"encoding/json"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/utils"
	"github.com/btcsuite/btcd/chaincfg"
)

type VerifyUTXOsRequestPayload struct {
	Address string                 `json:"address"`
	UTXOs   []types.UTXOIdentifier `json:"utxos"`
}

func parseRequestPayload(request *http.Request, maxUTXOs uint32, netParam *chaincfg.Params) (*VerifyUTXOsRequestPayload, *types.Error) {
	var payload VerifyUTXOsRequestPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, "invalid input format")
	}
	utxos := payload.UTXOs
	if len(utxos) == 0 {
		return nil, types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, "empty UTXO array")
	}

	if uint32(len(utxos)) > maxUTXOs {
		return nil, types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, "too many UTXOs in the request")
	}

	for _, utxo := range utxos {
		if !utils.IsValidTxHash(utxo.Txid) {
			return nil, types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, "invalid UTXO txid")
		} else if utxo.Vout < 0 {
			return nil, types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, "invalid UTXO vout")
		}
	}

	if _, err := utils.CheckBtcAddressType(payload.Address, netParam); err != nil {
		return nil, types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, err.Error())
	}
	return &payload, nil
}

func (h *Handler) VerifyUTXOs(request *http.Request) (*Result, *types.Error) {
	inputs, err := parseRequestPayload(request, h.Config.Assets.MaxUTXOs, h.Config.Server.BTCNetParam)
	if err != nil {
		return nil, err
	}

	results, err := h.Service.VerifyUTXOs(request.Context(), inputs.UTXOs, inputs.Address)
	if err != nil {
		return nil, err
	}

	return NewResult(results), nil
}
