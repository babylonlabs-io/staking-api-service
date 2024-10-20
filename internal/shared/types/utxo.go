package types

type UTXOIdentifier struct {
	Txid string `json:"txid"`
	Vout uint32 `json:"vout"`
}
