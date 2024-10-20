package utils

import (
	"encoding/base64"
	"encoding/hex"
	"regexp"

	bbntypes "github.com/babylonlabs-io/babylon/types"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// IsValidTxHash checks if the given string is a valid BTC transaction hash
// Note: it does not check the actual content of the hash.
func IsValidTxHash(txHash string) bool {
	// Check if the hash is valid
	_, err := chainhash.NewHashFromStr(txHash)
	return err == nil
}

// IsBase64Encoded checks if the given string is a valid Base64 encoded string.
// Note: it does not check the actual content of the string.
func IsBase64Encoded(s string) bool {
	// Check if the string length is a multiple of 4.
	if len(s)%4 != 0 {
		return false
	}

	// Regular expression to match valid Base64 characters.
	base64Regex := regexp.MustCompile(`^[a-zA-Z0-9+/]*={0,2}$`)
	if !base64Regex.MatchString(s) {
		return false
	}

	// Try to decode the string.
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// IsValidTxHex checks if the given string is a valid BTC transaction hex
// Note: it does not check the actual content of the transaction.
func IsValidTxHex(txHex string) bool {
	_, _, err := bbntypes.NewBTCTxFromHex(txHex)
	return err == nil
}

// IsValidSignatureFormat checks if the given string is a valid signature in hex format
// Note: it does not check the actual content of the signature.
func IsValidSignatureFormat(sigHex string) bool {
	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil {
		return false
	}
	_, err = schnorr.ParseSignature(sigBytes)
	return err == nil
}
