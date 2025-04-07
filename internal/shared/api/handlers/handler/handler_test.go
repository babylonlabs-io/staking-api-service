package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateBabylonAddress(t *testing.T) {
	t.Run("Valid addresses", func(t *testing.T) {
		addresses := []string{
			"bbn1cyqgpk0nlsutlm5ymkfpya30fqntanc8slpure",
			"bbn1rey7n439hgmzeqtd6s6636dcxzm389s9ay7dn7",
		}

		for _, addr := range addresses {
			err := ValidateBabylonAddress(addr)
			assert.NoError(t, err)
		}
	})
	t.Run("Invalid addresses", func(t *testing.T) {
		addresses := []string{
			"cosmos1t8y2z67p4a3cv8ef2l9w6k5j0m8q4r2a1s0d5f",
		}

		for _, addr := range addresses {
			err := ValidateBabylonAddress(addr)
			assert.Error(t, err)
		}
	})
}
