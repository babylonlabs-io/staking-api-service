package datagen

import (
	"bytes"
	"encoding/hex"
	"math/rand"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenRandomByteArray(r *rand.Rand, length uint64) []byte {
	newHeaderBytes := make([]byte, length)
	r.Read(newHeaderBytes)
	return newHeaderBytes
}

func RandomPk() (string, error) {
	fpPirvKey, err := btcec.NewPrivateKey()
	if err != nil {
		return "", err
	}
	fpPk := fpPirvKey.PubKey()
	return hex.EncodeToString(schnorr.SerializePubKey(fpPk)), nil
}

// RandomPostiveFloat64 generates a random float64 value greater than 0.
func RandomPostiveFloat64(r *rand.Rand) float64 {
	for {
		f := r.Float64() // Generate a random float64
		if f > 0 {
			return f
		}
		// If f is 0 (extremely rare), regenerate
	}
}

// RandomPositiveInt generates a random positive integer from 1 to max.
func RandomPositiveInt(r *rand.Rand, max int) int {
	// Generate a random number from 1 to max (inclusive)
	return r.Intn(max) + 1
}

// RandomString generates a random alphanumeric string of length n.
func RandomString(r *rand.Rand, n int) string {
	result := make([]byte, n)
	letterLen := len(letters)
	for i := range result {
		num := r.Int() % letterLen
		result[i] = letters[num]
	}
	return string(result)
}

// RandomAmount generates a random BTC amount from 0.1 to 10000
// the returned value is in satoshis
func RandomAmount(r *rand.Rand) int64 {
	// Generate a random value range from 0.1 to 10000 BTC
	randomBTC := r.Float64()*(9999.9-0.1) + 0.1
	// convert to satoshi
	return int64(randomBTC*1e8) + 1
}

// GenerateRandomTx generates a random transaction with random values for each field.
func GenerateRandomTx(
	r *rand.Rand,
	options *struct{ DisableRbf bool },
) (*wire.MsgTx, string, error) {
	sequence := r.Uint32()
	if options != nil && options.DisableRbf {
		sequence = wire.MaxTxInSequenceNum
	}
	tx := &wire.MsgTx{
		Version: 1,
		TxIn: []*wire.TxIn{
			{
				PreviousOutPoint: wire.OutPoint{
					Hash:  chainhash.HashH(GenRandomByteArray(r, 10)),
					Index: r.Uint32(),
				},
				SignatureScript: []byte{},
				Sequence:        sequence,
			},
		},
		TxOut: []*wire.TxOut{
			{
				Value:    int64(r.Int31()),
				PkScript: GenRandomByteArray(r, 80),
			},
		},
		LockTime: 0,
	}
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return nil, "", err
	}
	txHex := hex.EncodeToString(buf.Bytes())

	return tx, txHex, nil
}

func RandomDelegationState(r *rand.Rand) types.DelegationState {
	states := []types.DelegationState{types.Active, types.UnbondingRequested, types.Unbonding, types.Unbonded, types.Withdrawn}
	return states[r.Intn(len(states))]
}
