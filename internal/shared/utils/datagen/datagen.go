package datagen

import (
	"bytes"
	"encoding/hex"
	"math/rand"
	"time"

	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
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

// GenerateRandomTxWithOutput generates a random transaction with random values
// for each field.
func RandomBytes(r *rand.Rand, n uint64) ([]byte, string) {
	randomBytes := GenRandomByteArray(r, n)
	return randomBytes, hex.EncodeToString(randomBytes)
}

// GenerateRandomTimestamp generates a random timestamp before the specified timestamp.
// If beforeTimestamp is 0, then the current time is used.
func GenerateRandomTimestamp(afterTimestamp, beforeTimestamp int64) int64 {
	timeNow := time.Now().Unix()
	if beforeTimestamp == 0 && afterTimestamp == 0 {
		return timeNow
	}
	if beforeTimestamp == 0 {
		return afterTimestamp + rand.Int63n(timeNow-afterTimestamp)
	} else if afterTimestamp == 0 {
		// Generate a reasonable timestamp between 1 second to 6 months in the past
		sixMonthsInSeconds := int64(6 * 30 * 24 * 60 * 60)
		return beforeTimestamp - rand.Int63n(sixMonthsInSeconds)
	}
	return afterTimestamp + rand.Int63n(beforeTimestamp-afterTimestamp)
}

func RandomFinalityProviderState(r *rand.Rand) types.FinalityProviderQueryingState {
	states := []types.FinalityProviderQueryingState{types.FinalityProviderStateActive, types.FinalityProviderStateStandby}
	return states[r.Intn(len(states))]
}

func GenerateRandomBTCParams(r *rand.Rand) indexertypes.BtcCheckpointParams {
	return indexertypes.BtcCheckpointParams{
		Version:              uint32(r.Intn(10)),
		BtcConfirmationDepth: uint64(r.Intn(10)),
	}
}

func RandomDelegationState(r *rand.Rand) types.DelegationState {
	states := []types.DelegationState{types.Active, types.UnbondingRequested, types.Unbonding, types.Unbonded, types.Withdrawn}
	return states[r.Intn(len(states))]
}

func RandomTransactionInfo(r *rand.Rand) types.TransactionInfo {
	_, txHex, _ := GenerateRandomTx(r, nil)
	return types.TransactionInfo{
		TxHex:       txHex,
		OutputIndex: r.Intn(100),
	}
}
