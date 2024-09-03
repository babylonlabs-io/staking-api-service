package utilstest

import (
	"testing"

	"github.com/babylonlabs-io/staking-api-service/internal/utils"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/assert"
)

func TestDeriveAddressesFromNoCoordPk(t *testing.T) {
	// Given a known BTC public key hex string (32 bytes X coordinate)
	pkHex := "30bb400d3ef60a5bb66a3f5d9e0e870ccbf8ae1a4ab2263a9fabf90adf94c70a"
	expectedTaprootAddress := "bc1p89k3uz7fdt58gl3vtxqvfxcsgh0t923qfxyuw8l5qdz70fsxzzqq35fjjt"
	expectedEvenAddress := "bc1qem92n3xk2rm72mua7jq66m700m3r2ama60mc35"
	expectedOddAddress := "bc1q032et33x53y97ersj3k32dytfuh2lef3ewy4z2"

	addresses, err := utils.DeriveAddressesFromNoCoordPk(pkHex, &chaincfg.MainNetParams)

	// Then assert that there is no error
	assert.NoError(t, err)

	// And assert that the derived addresses match the expected values
	assert.Equal(t, expectedTaprootAddress, addresses.Taproot)
	assert.Equal(t, expectedEvenAddress, addresses.NativeSegwitEven)
	assert.Equal(t, expectedOddAddress, addresses.NativeSegwitOdd)
}
