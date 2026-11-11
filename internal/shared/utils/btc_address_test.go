package utils

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/assert"
)

func TestIsValidBtcAddress(t *testing.T) {
	mainnet := &chaincfg.MainNetParams
	signet := &chaincfg.SigNetParams

	cases := []struct {
		name    string
		address string
		params  *chaincfg.Params
		want    bool
	}{
		{"mainnet P2PKH", "1BoatSLRHtKNngkdXEeobR76b53LETtpyT", mainnet, true},
		{"mainnet P2SH", "3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy", mainnet, true},
		{"mainnet P2WPKH", "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq", mainnet, true},
		{"mainnet P2TR", "bc1p5d7rjq7g6rdk2yhzks9smlaqtedr4dekq08ge8ztwac72sfr9rusxg3297", mainnet, true},
		{"signet P2WPKH", "tb1qw508d6qejxtdg4y5r3zarvary0c5xw7kxpjzsx", signet, true},
		{"signet address on mainnet", "tb1qw508d6qejxtdg4y5r3zarvary0c5xw7kxpjzsx", mainnet, false},
		{"mainnet address on signet", "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq", signet, false},
		{"empty", "", mainnet, false},
		{"garbage", "invalid", mainnet, false},
		{"path traversal", "x/../../", mainnet, false},
		{"file path", "../../etc/passwd", mainnet, false},
		{"url", "https://example.com/screen", mainnet, false},
		{"valid address with suffix", "1BoatSLRHtKNngkdXEeobR76b53LETtpyT/../x", mainnet, false},
		{"valid address with query", "1BoatSLRHtKNngkdXEeobR76b53LETtpyT?x=1", mainnet, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsValidBtcAddress(tc.address, tc.params))
		})
	}
}
