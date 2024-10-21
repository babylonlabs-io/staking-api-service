package types

type TransactionInfo struct {
	TxHex       string `json:"tx_hex"`
	OutputIndex int    `json:"output_index"`
}
